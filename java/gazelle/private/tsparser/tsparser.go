// Package tsparser implements Java metadata extraction using gotreesitter.
// It is a pure-Go replacement for the gRPC-based javaparser that requires
// a Java subprocess.
package tsparser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
	"github.com/rs/zerolog"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/parser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_multiset"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

var (
	tsJavaLang = grammars.DetectLanguage("Test.java").Language()
	// ParserPool is concurrency-safe and resets parser state between uses.
	tsJavaParserPool = gotreesitter.NewParserPool(tsJavaLang)
)

// Runner implements Java parsing using gotreesitter (pure Go, no subprocess).
type Runner struct {
	logger   zerolog.Logger
	repoRoot string
}

// NewRunner creates a tree-sitter-based Java parser.
// No subprocess or JDK required.
func NewRunner(logger zerolog.Logger, repoRoot string) *Runner {
	return &Runner{
		logger:   logger.With().Str("_c", "javaparser-ts").Logger(),
		repoRoot: repoRoot,
	}
}

func (r *Runner) Shutdown() {}

func (r *Runner) ParsePackage(ctx context.Context, in *parser.ParsePackageRequest) (*java.Package, error) {
	defer func(t time.Time) {
		r.logger.Debug().
			Str("duration", time.Since(t).String()).
			Str("rel", in.Rel).
			Msg("parse package done")
	}(time.Now())

	importedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	exportedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	importedPackages := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	mains := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	perClassMetadata := make(map[string]java.PerClassMetadata)

	var packageName types.PackageName

	for _, filename := range in.Files {
		info, err := r.parseJavaFile(in.Rel, filename)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", filename, err)
		}

		if packageName.Name == "" && info.packageName != "" {
			packageName = types.NewPackageName(info.packageName)
		}

		for _, imp := range info.importedClasses {
			cn, err := types.ParseClassName(imp)
			if err != nil {
				r.logger.Warn().Str("import", imp).Err(err).Msg("skipping unparseable import")
				continue
			}
			importedClasses.Add(*cn)
		}
		for _, pkg := range info.importedPackages {
			importedPackages.Add(types.NewPackageName(pkg))
		}

		for _, cls := range info.exportedClasses {
			fqn := cls
			if packageName.Name != "" {
				fqn = packageName.Name + "." + cls
			}
			cn, err := types.ParseClassName(fqn)
			if err != nil {
				r.logger.Warn().Str("class", fqn).Err(err).Msg("skipping unparseable exported class")
				continue
			}
			exportedClasses.Add(*cn)
		}

		for _, cls := range info.mainClasses {
			mains.Add(types.NewClassName(packageName, cls))
		}

		for className, meta := range info.perClassMetadata {
			perClassMetadata[className] = meta
		}
	}

	return &java.Package{
		Name:                                   packageName,
		ImportedClasses:                        importedClasses,
		ExportedClasses:                        exportedClasses,
		ImportedPackagesWithoutSpecificClasses: importedPackages,
		Mains:                                  mains,
		Files:                                  sorted_set.NewSortedSet(in.Files),
		TestPackage:                            java.IsTestPackage(in.Rel),
		PerClassMetadata:                       perClassMetadata,
	}, nil
}

// ---------------------------------------------------------------------------
// Per-file extraction
// ---------------------------------------------------------------------------

type javaFileInfo struct {
	packageName      string
	importedClasses  []string // FQN of imported classes
	importedPackages []string // package names from wildcard imports
	exportedClasses  []string // bare names of public top-level types
	mainClasses      []string // bare names of classes containing main()
	perClassMetadata map[string]java.PerClassMetadata
}

func (r *Runner) parseJavaFile(rel, filename string) (*javaFileInfo, error) {
	path := filepath.Join(r.repoRoot, rel, filename)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tree, err := r.parseJavaContent(content)
	if err != nil {
		return nil, err
	}
	defer tree.Release()
	root := tree.RootNode()

	info := &javaFileInfo{
		perClassMetadata: make(map[string]java.PerClassMetadata),
	}
	// Import map for resolving simple annotation names to FQNs.
	importMap := make(map[string]string)

	for i := 0; i < root.NamedChildCount(); i++ {
		child := root.NamedChild(i)
		switch child.Type(tsJavaLang) {
		case "package_declaration":
			info.packageName = extractPackageName(child, content)
		case "import_declaration":
			extractImport(child, content, info, importMap)
		case "class_declaration", "interface_declaration", "enum_declaration":
			extractTypeDecl(child, content, info, importMap)
		}
	}

	return info, nil
}

func (r *Runner) parseJavaContent(content []byte) (*gotreesitter.Tree, error) {
	tree, err := tsJavaParserPool.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}
	if tree == nil || tree.RootNode() == nil {
		return nil, fmt.Errorf("parse: nil root")
	}

	// Fallback to the Java token-source path when the default parse tree has
	// syntax errors. This guards metadata extraction against parser regressions
	// in one parse mode while preserving the fast path for healthy trees.
	if !tree.RootNode().HasError() {
		return tree, nil
	}

	ts, err := grammars.NewJavaTokenSource(content, tsJavaLang)
	if err != nil {
		r.logger.Debug().Err(err).Msg("java token-source fallback unavailable")
		return tree, nil
	}

	fallbackTree, err := tsJavaParserPool.ParseWithTokenSource(content, ts)
	if err != nil {
		if fallbackTree != nil {
			fallbackTree.Release()
		}
		r.logger.Debug().Err(err).Msg("java token-source fallback parse failed")
		return tree, nil
	}
	if fallbackTree == nil || fallbackTree.RootNode() == nil {
		if fallbackTree != nil {
			fallbackTree.Release()
		}
		r.logger.Debug().Msg("java token-source fallback returned nil root")
		return tree, nil
	}
	if fallbackTree.RootNode().HasError() {
		fallbackTree.Release()
		return tree, nil
	}

	tree.Release()
	r.logger.Debug().Msg("java parse recovered by token-source fallback")
	return fallbackTree, nil
}

// ---------------------------------------------------------------------------
// Extraction helpers
// ---------------------------------------------------------------------------

func extractPackageName(node *gotreesitter.Node, content []byte) string {
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		typ := child.Type(tsJavaLang)
		if typ == "scoped_identifier" || typ == "identifier" {
			return child.Text(content)
		}
	}
	return ""
}

func extractImport(node *gotreesitter.Node, content []byte, info *javaFileInfo, importMap map[string]string) {
	// Detect "static" keyword among anonymous children.
	isStatic := false
	for i := 0; i < node.ChildCount(); i++ {
		ch := node.Child(i)
		if !ch.IsNamed() && ch.Text(content) == "static" {
			isStatic = true
			break
		}
	}

	var importPath string
	hasAsterisk := false
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		typ := child.Type(tsJavaLang)
		if typ == "scoped_identifier" || typ == "identifier" {
			importPath = child.Text(content)
		} else if typ == "asterisk" {
			hasAsterisk = true
		}
	}
	if importPath == "" {
		return
	}

	if isStatic {
		if hasAsterisk {
			// import static com.example.Foo.* → depends on class Foo
			info.importedClasses = append(info.importedClasses, importPath)
			addToImportMap(importMap, importPath)
		} else {
			// import static com.example.Foo.bar → strip member, keep class
			if dot := strings.LastIndex(importPath, "."); dot >= 0 {
				className := importPath[:dot]
				info.importedClasses = append(info.importedClasses, className)
				addToImportMap(importMap, className)
			}
		}
	} else {
		if hasAsterisk {
			// import com.example.* → wildcard package import
			info.importedPackages = append(info.importedPackages, importPath)
		} else {
			// import com.example.Foo → specific class import
			info.importedClasses = append(info.importedClasses, importPath)
			addToImportMap(importMap, importPath)
		}
	}
}

func addToImportMap(importMap map[string]string, fqn string) {
	if dot := strings.LastIndex(fqn, "."); dot >= 0 {
		importMap[fqn[dot+1:]] = fqn
	} else {
		importMap[fqn] = fqn
	}
}

func extractTypeDecl(node *gotreesitter.Node, content []byte, info *javaFileInfo, importMap map[string]string) {
	nameNode := node.ChildByFieldName("name", tsJavaLang)
	if nameNode == nil {
		return
	}
	className := nameNode.Text(content)

	if hasModifier(node, content, "public") {
		info.exportedClasses = append(info.exportedClasses, className)
	}

	// Class-level annotations.
	classAnns := extractAnnotationNames(node, content, importMap, info.packageName)
	classAnnSet := sorted_set.NewSortedSetFn(nil, types.ClassNameLess)
	for _, ann := range classAnns {
		if cn, err := types.ParseClassName(ann); err == nil {
			classAnnSet.Add(*cn)
		}
	}

	methodAnns := sorted_multiset.NewSortedMultiSetFn[string, types.ClassName](types.ClassNameLess)
	fieldAnns := sorted_multiset.NewSortedMultiSetFn[string, types.ClassName](types.ClassNameLess)

	bodyNode := node.ChildByFieldName("body", tsJavaLang)
	if bodyNode != nil {
		for i := 0; i < bodyNode.NamedChildCount(); i++ {
			member := bodyNode.NamedChild(i)
			memberType := member.Type(tsJavaLang)

			switch memberType {
			case "method_declaration":
				methodName := findIdentifier(member, content)

				if node.Type(tsJavaLang) == "class_declaration" && methodName == "main" && isMainMethod(member, content) {
					info.mainClasses = append(info.mainClasses, className)
				}

				for _, ann := range extractAnnotationNames(member, content, importMap, info.packageName) {
					if cn, err := types.ParseClassName(ann); err == nil {
						methodAnns.Add(methodName, *cn)
					}
				}

			case "field_declaration":
				fieldNames := extractFieldNames(member, content)
				annotations := extractAnnotationNames(member, content, importMap, info.packageName)
				for _, fieldName := range fieldNames {
					for _, ann := range annotations {
						if cn, err := types.ParseClassName(ann); err == nil {
							fieldAnns.Add(fieldName, *cn)
						}
					}
				}
			}
		}
	}

	if classAnnSet.Len() > 0 || len(methodAnns.Keys()) > 0 || len(fieldAnns.Keys()) > 0 {
		info.perClassMetadata[className] = java.PerClassMetadata{
			AnnotationClassNames:       classAnnSet,
			MethodAnnotationClassNames: methodAnns,
			FieldAnnotationClassNames:  fieldAnns,
		}
	}
}

func hasModifier(node *gotreesitter.Node, content []byte, modifier string) bool {
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child.Type(tsJavaLang) == "modifiers" {
			for j := 0; j < child.ChildCount(); j++ {
				token := child.Child(j)
				if !token.IsNamed() && token.Text(content) == modifier {
					return true
				}
			}
			return false
		}
	}
	return false
}

func extractAnnotationNames(node *gotreesitter.Node, content []byte, importMap map[string]string, packageName string) []string {
	var annotations []string
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		switch child.Type(tsJavaLang) {
		case "marker_annotation", "annotation":
			if name := resolveAnnotationName(child, content, importMap, packageName); name != "" {
				annotations = append(annotations, name)
			}
		case "modifiers":
			for j := 0; j < child.NamedChildCount(); j++ {
				annNode := child.NamedChild(j)
				annType := annNode.Type(tsJavaLang)
				if annType != "marker_annotation" && annType != "annotation" {
					continue
				}
				if name := resolveAnnotationName(annNode, content, importMap, packageName); name != "" {
					annotations = append(annotations, name)
				}
			}
		}
	}
	return annotations
}

func resolveAnnotationName(node *gotreesitter.Node, content []byte, importMap map[string]string, packageName string) string {
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		typ := child.Type(tsJavaLang)
		switch typ {
		case "identifier":
			simple := child.Text(content)
			if fqn, ok := importMap[simple]; ok {
				return fqn
			}
			if isJavaLangType(simple) {
				return "java.lang." + simple
			}
			if packageName != "" {
				return packageName + "." + simple
			}
			return simple
		case "scoped_identifier":
			return child.Text(content)
		}
	}
	return ""
}

// isJavaLangType returns true for annotation types in java.lang (auto-imported).
func isJavaLangType(name string) bool {
	switch name {
	case "Override", "Deprecated", "SuppressWarnings", "SafeVarargs", "FunctionalInterface":
		return true
	}
	return false
}

func isMainMethod(node *gotreesitter.Node, content []byte) bool {
	if !hasModifier(node, content, "public") || !hasModifier(node, content, "static") {
		return false
	}
	// Field map entries may be incomplete for method_declaration,
	// so walk children by type instead of using ChildByFieldName.
	var hasVoidReturn bool
	var hasStringArrayParam bool
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		switch child.Type(tsJavaLang) {
		case "void_type":
			hasVoidReturn = true
		case "formal_parameters":
			paramText := child.Text(content)
			hasStringArrayParam = isSingleStringMainParam(paramText)
		}
	}
	return hasVoidReturn && hasStringArrayParam
}

func isSingleStringMainParam(paramText string) bool {
	paramText = strings.TrimSpace(paramText)
	if len(paramText) < 2 || paramText[0] != '(' || paramText[len(paramText)-1] != ')' {
		return false
	}
	inner := strings.TrimSpace(paramText[1 : len(paramText)-1])
	if inner == "" {
		return false
	}
	// main should have exactly one parameter; reject additional params.
	if strings.Contains(inner, ",") {
		return false
	}

	noSpace := strings.ReplaceAll(inner, " ", "")
	return strings.Contains(noSpace, "String[]") || strings.Contains(noSpace, "String...")
}

// findIdentifier returns the text of the first direct identifier child.
// For method_declaration, this is the method name.
func findIdentifier(node *gotreesitter.Node, content []byte) string {
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child.Type(tsJavaLang) == "identifier" {
			return child.Text(content)
		}
	}
	return ""
}

func extractFieldNames(node *gotreesitter.Node, content []byte) []string {
	var fieldNames []string
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child.Type(tsJavaLang) == "variable_declarator" {
			if fieldName := findIdentifier(child, content); fieldName != "" {
				// variable_declarator's first identifier child is the field name.
				fieldNames = append(fieldNames, fieldName)
			}
		}
	}
	return fieldNames
}

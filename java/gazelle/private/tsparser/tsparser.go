// Package tsparser implements Java metadata extraction using gotreesitter.
// It is a pure-Go replacement for the gRPC-based javaparser that requires
// a Java subprocess.
package tsparser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	importedClassNames := make(map[string]struct{})
	exportedClassNames := make(map[string]struct{})
	importedPackageNames := make(map[string]struct{})
	mainClassNames := make(map[string]struct{})
	perClassMetadata := make(map[string]java.PerClassMetadata)

	var packageName types.PackageName
	packageNames := map[string]struct{}{}

	for _, filename := range in.Files {
		info, err := r.parseJavaFile(in.Rel, filename)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", filename, err)
		}

		if info.packageName != "" {
			packageNames[info.packageName] = struct{}{}
		}
		if packageName.Name == "" && info.packageName != "" {
			packageName = types.NewPackageName(info.packageName)
		}

		for _, imp := range info.importedClasses {
			importedClassNames[imp] = struct{}{}
		}
		for _, pkg := range info.importedPackages {
			importedPackageNames[pkg] = struct{}{}
		}

		for _, cls := range info.exportedClasses {
			exportedClassNames[cls] = struct{}{}
		}

		for _, cls := range info.mainClasses {
			mainClassName := types.NewClassName(packageName, cls)
			mainClassNames[mainClassName.FullyQualifiedClassName()] = struct{}{}
		}

		for className, meta := range info.perClassMetadata {
			perClassMetadata[className] = meta
		}
	}
	if len(packageNames) > 1 {
		names := make([]string, 0, len(packageNames))
		for name := range packageNames {
			names = append(names, name)
		}
		sort.Strings(names)
		return nil, fmt.Errorf("InvalidArgument: Expected exactly one java package, but saw %d: %s", len(names), strings.Join(names, ", "))
	}

	importedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for imp := range importedClassNames {
		cn, err := types.ParseClassName(imp)
		if err != nil {
			r.logger.Warn().Str("import", imp).Err(err).Msg("skipping unparseable import")
			continue
		}
		importedClasses.Add(*cn)
	}

	exportedClasses := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for cls := range exportedClassNames {
		cn, err := types.ParseClassName(cls)
		if err != nil {
			r.logger.Warn().Str("class", cls).Err(err).Msg("skipping unparseable exported class")
			continue
		}
		exportedClasses.Add(*cn)
	}

	importedPackages := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	for pkg := range importedPackageNames {
		importedPackages.Add(types.NewPackageName(pkg))
	}

	mains := sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess)
	for cls := range mainClassNames {
		cn, err := types.ParseClassName(cls)
		if err != nil {
			r.logger.Warn().Str("class", cls).Err(err).Msg("skipping unparseable main class")
			continue
		}
		mains.Add(*cn)
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
	exportedClasses  []string // FQN of public API dependency types
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
		}
	}

	localClassNames := collectLocalClassNames(tree, content)
	for i := 0; i < root.NamedChildCount(); i++ {
		child := root.NamedChild(i)
		switch child.Type(tsJavaLang) {
		case "class_declaration", "interface_declaration", "enum_declaration":
			extractTypeDecl(child, content, info, importMap, nil, localClassNames, nil)
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

func collectLocalClassNames(tree *gotreesitter.Tree, content []byte) map[string]struct{} {
	names := map[string]struct{}{}
	cursor := gotreesitter.NewTreeCursorFromTree(tree)
	for {
		node := cursor.CurrentNode()
		if node == nil {
			return names
		}
		switch node.Type(tsJavaLang) {
		case "class_declaration", "interface_declaration", "enum_declaration":
			if nameNode := node.ChildByFieldName("name", tsJavaLang); nameNode != nil {
				names[nameNode.Text(content)] = struct{}{}
			}
		}

		if cursor.GotoFirstNamedChild() {
			continue
		}
		for {
			if cursor.GotoNextNamedSibling() {
				break
			}
			if !cursor.GotoParent() {
				return names
			}
		}
	}
}

func extractTypeDecl(node *gotreesitter.Node, content []byte, info *javaFileInfo, importMap map[string]string, parents []string, localClassNames map[string]struct{}, inheritedTypeParams map[string]struct{}) {
	nameNode := node.ChildByFieldName("name", tsJavaLang)
	if nameNode == nil {
		return
	}
	className := nameNode.Text(content)
	nestedNames := append(append([]string{}, parents...), className)
	classFQN := qualifyNestedClassName(info.packageName, nestedNames)

	typeParams := extendTypeParameters(inheritedTypeParams, node, content)
	for _, ref := range directTypeRefs(node, content, importMap, info.packageName, localClassNames, typeParams, true) {
		info.importedClasses = append(info.importedClasses, ref)
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
			case "class_declaration", "interface_declaration", "enum_declaration":
				extractTypeDecl(member, content, info, importMap, nestedNames, localClassNames, typeParams)

			case "method_declaration":
				methodName := findIdentifier(member, content)
				methodTypeParams := extendTypeParameters(typeParams, member, content)
				for _, ref := range directTypeRefs(member, content, importMap, info.packageName, localClassNames, methodTypeParams, true) {
					info.importedClasses = append(info.importedClasses, ref)
				}
				if body := methodBodyNode(member); body != nil {
					for _, ref := range usedTypeRefs(body, content, importMap, info.packageName, localClassNames, methodTypeParams) {
						info.importedClasses = append(info.importedClasses, ref)
					}
				}
				if !hasModifier(member, content, "private") {
					if returnType := methodReturnTypeNode(member); returnType != nil && returnType.Type(tsJavaLang) != "void_type" {
						for _, ref := range typeRefsFromTypeNode(returnType, content, importMap, info.packageName, localClassNames, methodTypeParams, false) {
							info.exportedClasses = append(info.exportedClasses, ref)
						}
					}
				}

				if node.Type(tsJavaLang) == "class_declaration" && methodName == "main" && isMainMethod(member, content) {
					info.mainClasses = append(info.mainClasses, className)
				}

				for _, ann := range extractAnnotationNames(member, content, importMap, info.packageName) {
					if cn, err := types.ParseClassName(ann); err == nil {
						methodAnns.Add(methodName, *cn)
					}
				}

			case "field_declaration":
				for _, ref := range directTypeRefs(member, content, importMap, info.packageName, localClassNames, typeParams, true) {
					info.importedClasses = append(info.importedClasses, ref)
				}
				fieldNames := extractFieldNames(member, content)
				annotations := extractAnnotationNames(member, content, importMap, info.packageName)
				for _, fieldName := range fieldNames {
					for _, ann := range annotations {
						if cn, err := types.ParseClassName(ann); err == nil {
							fieldAnns.Add(fieldName, *cn)
						}
					}
				}
			default:
				for _, ref := range usedTypeRefs(member, content, importMap, info.packageName, localClassNames, typeParams) {
					info.importedClasses = append(info.importedClasses, ref)
				}
			}
		}
	}

	if classAnnSet.Len() > 0 || len(methodAnns.Keys()) > 0 || len(fieldAnns.Keys()) > 0 {
		info.perClassMetadata[classFQN] = java.PerClassMetadata{
			AnnotationClassNames:       classAnnSet,
			MethodAnnotationClassNames: methodAnns,
			FieldAnnotationClassNames:  fieldAnns,
		}
	}
}

func qualifyNestedClassName(packageName string, names []string) string {
	parts := make([]string, 0, len(names)+1)
	if packageName != "" {
		parts = append(parts, packageName)
	}
	parts = append(parts, names...)
	return strings.Join(parts, ".")
}

func cloneStringSet(in map[string]struct{}) map[string]struct{} {
	if len(in) == 0 {
		return map[string]struct{}{}
	}
	out := map[string]struct{}{}
	for k := range in {
		out[k] = struct{}{}
	}
	return out
}

func extendTypeParameters(in map[string]struct{}, node *gotreesitter.Node, content []byte) map[string]struct{} {
	var out map[string]struct{}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child.Type(tsJavaLang) != "type_parameters" {
			continue
		}
		for j := 0; j < child.NamedChildCount(); j++ {
			typeParam := child.NamedChild(j)
			if typeParam.Type(tsJavaLang) != "type_parameter" {
				continue
			}
			name := typeParameterName(typeParam, content)
			if name != "" {
				if out == nil {
					out = cloneStringSet(in)
				}
				out[name] = struct{}{}
			}
		}
	}
	if out != nil {
		return out
	}
	return in
}

func typeParameterName(node *gotreesitter.Node, content []byte) string {
	if name := node.ChildByFieldName("name", tsJavaLang); name != nil {
		return name.Text(content)
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if child.Type(tsJavaLang) == "type_identifier" {
			return child.Text(content)
		}
	}
	return ""
}

func methodReturnTypeNode(node *gotreesitter.Node) *gotreesitter.Node {
	if typ := node.ChildByFieldName("type", tsJavaLang); typ != nil {
		if isTypeNode(typ) || typ.Type(tsJavaLang) == "void_type" {
			return typ
		}
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		if isTypeNode(child) {
			return child
		}
	}
	return nil
}

func methodBodyNode(node *gotreesitter.Node) *gotreesitter.Node {
	if body := node.ChildByFieldName("body", tsJavaLang); body != nil {
		return body
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		switch child.Type(tsJavaLang) {
		case "block", "constructor_body":
			return child
		}
	}
	return nil
}

func directTypeRefs(node *gotreesitter.Node, content []byte, importMap map[string]string, packageName string, localClassNames, typeParams map[string]struct{}, includeParameterizedBase bool) []string {
	refs := make([]string, 0, 4)
	for i := 0; i < node.NamedChildCount(); i++ {
		child := node.NamedChild(i)
		switch child.Type(tsJavaLang) {
		case "class_body", "interface_body", "enum_body", "constructor_body", "block":
			continue
		}
		appendTypeRefsFromAnyNode(&refs, child, content, importMap, packageName, localClassNames, typeParams, includeParameterizedBase)
	}
	return refs
}

func usedTypeRefs(node *gotreesitter.Node, content []byte, importMap map[string]string, packageName string, localClassNames, typeParams map[string]struct{}) []string {
	refs := make([]string, 0, 4)
	appendTypeRefsFromAnyNode(&refs, node, content, importMap, packageName, localClassNames, typeParams, true)
	return refs
}

func typeRefsFromAnyNode(node *gotreesitter.Node, content []byte, importMap map[string]string, packageName string, localClassNames, typeParams map[string]struct{}, includeParameterizedBase bool) []string {
	refs := make([]string, 0, 4)
	appendTypeRefsFromAnyNode(&refs, node, content, importMap, packageName, localClassNames, typeParams, includeParameterizedBase)
	return refs
}

func appendTypeRefsFromAnyNode(refs *[]string, node *gotreesitter.Node, content []byte, importMap map[string]string, packageName string, localClassNames, typeParams map[string]struct{}, includeParameterizedBase bool) {
	if isTypeNode(node) {
		appendTypeRefsFromTypeNode(refs, node, content, importMap, packageName, localClassNames, typeParams, includeParameterizedBase)
		return
	}
	if node.Type(tsJavaLang) == "type_parameter" {
		name := typeParameterName(node, content)
		for i := 0; i < node.NamedChildCount(); i++ {
			child := node.NamedChild(i)
			if child.Text(content) == name {
				continue
			}
			appendTypeRefsFromAnyNode(refs, child, content, importMap, packageName, localClassNames, typeParams, includeParameterizedBase)
		}
		return
	}
	for i := 0; i < node.NamedChildCount(); i++ {
		appendTypeRefsFromAnyNode(refs, node.NamedChild(i), content, importMap, packageName, localClassNames, typeParams, includeParameterizedBase)
	}
}

func typeRefsFromTypeNode(node *gotreesitter.Node, content []byte, importMap map[string]string, packageName string, localClassNames, typeParams map[string]struct{}, includeParameterizedBase bool) []string {
	refs := make([]string, 0, 2)
	appendTypeRefsFromTypeNode(&refs, node, content, importMap, packageName, localClassNames, typeParams, includeParameterizedBase)
	return refs
}

func appendTypeRefsFromTypeNode(refs *[]string, node *gotreesitter.Node, content []byte, importMap map[string]string, packageName string, localClassNames, typeParams map[string]struct{}, includeParameterizedBase bool) {
	switch node.Type(tsJavaLang) {
	case "type_identifier", "scoped_type_identifier":
		if ref := resolveTypeName(node.Text(content), importMap, packageName, localClassNames, typeParams); ref != "" {
			*refs = append(*refs, ref)
		}
	case "generic_type":
		base := node.ChildByFieldName("type", tsJavaLang)
		if includeParameterizedBase && base != nil {
			appendTypeRefsFromTypeNode(refs, base, content, importMap, packageName, localClassNames, typeParams, true)
		}
		for i := 0; i < node.NamedChildCount(); i++ {
			child := node.NamedChild(i)
			if sameNode(child, base) {
				continue
			}
			appendTypeRefsFromAnyNode(refs, child, content, importMap, packageName, localClassNames, typeParams, true)
		}
	case "array_type":
		for i := 0; i < node.NamedChildCount(); i++ {
			appendTypeRefsFromAnyNode(refs, node.NamedChild(i), content, importMap, packageName, localClassNames, typeParams, includeParameterizedBase)
		}
	default:
		for i := 0; i < node.NamedChildCount(); i++ {
			appendTypeRefsFromAnyNode(refs, node.NamedChild(i), content, importMap, packageName, localClassNames, typeParams, includeParameterizedBase)
		}
	}
}

func sameNode(a, b *gotreesitter.Node) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Symbol() == b.Symbol() && a.StartByte() == b.StartByte() && a.EndByte() == b.EndByte()
}

func isTypeNode(node *gotreesitter.Node) bool {
	switch node.Type(tsJavaLang) {
	case "type_identifier", "scoped_type_identifier", "generic_type", "array_type", "wildcard":
		return true
	}
	return false
}

func resolveTypeName(typeName string, importMap map[string]string, packageName string, localClassNames, typeParams map[string]struct{}) string {
	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		return ""
	}
	parts := strings.Split(typeName, ".")
	if fqn, ok := importMap[parts[0]]; ok {
		return fqn
	}
	if len(parts) > 1 {
		return typeName
	}
	if _, ok := localClassNames[typeName]; ok {
		return ""
	}
	if _, ok := typeParams[typeName]; ok {
		return ""
	}
	if isJavaLangType(typeName) {
		return ""
	}
	if packageName == "" {
		return ""
	}
	return packageName + "." + typeName
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
			for _, token := range strings.Fields(child.Text(content)) {
				if token == modifier {
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
			return simple
		case "scoped_identifier":
			return child.Text(content)
		}
	}
	return ""
}

func isJavaLangType(name string) bool {
	_, ok := javaLangTypes[name]
	return ok
}

var javaLangTypes = map[string]struct{}{
	"Boolean":                         {},
	"Byte":                            {},
	"Character":                       {},
	"Double":                          {},
	"Float":                           {},
	"Integer":                         {},
	"Long":                            {},
	"Short":                           {},
	"Void":                            {},
	"CharSequence":                    {},
	"Class":                           {},
	"ClassLoader":                     {},
	"Comparable":                      {},
	"Enum":                            {},
	"Iterable":                        {},
	"Math":                            {},
	"Number":                          {},
	"Object":                          {},
	"Package":                         {},
	"Process":                         {},
	"ProcessBuilder":                  {},
	"Record":                          {},
	"Runtime":                         {},
	"SecurityManager":                 {},
	"StackTraceElement":               {},
	"StrictMath":                      {},
	"String":                          {},
	"StringBuffer":                    {},
	"StringBuilder":                   {},
	"System":                          {},
	"Thread":                          {},
	"ThreadGroup":                     {},
	"ThreadLocal":                     {},
	"Appendable":                      {},
	"AutoCloseable":                   {},
	"Cloneable":                       {},
	"Readable":                        {},
	"Runnable":                        {},
	"Throwable":                       {},
	"Error":                           {},
	"Exception":                       {},
	"RuntimeException":                {},
	"ArithmeticException":             {},
	"ArrayIndexOutOfBoundsException":  {},
	"ArrayStoreException":             {},
	"ClassCastException":              {},
	"ClassNotFoundException":          {},
	"CloneNotSupportedException":      {},
	"EnumConstantNotPresentException": {},
	"IllegalAccessException":          {},
	"IllegalArgumentException":        {},
	"IllegalMonitorStateException":    {},
	"IllegalStateException":           {},
	"IllegalThreadStateException":     {},
	"IndexOutOfBoundsException":       {},
	"InstantiationException":          {},
	"InterruptedException":            {},
	"NegativeArraySizeException":      {},
	"NoSuchFieldException":            {},
	"NoSuchMethodException":           {},
	"NullPointerException":            {},
	"NumberFormatException":           {},
	"ReflectiveOperationException":    {},
	"SecurityException":               {},
	"StringIndexOutOfBoundsException": {},
	"TypeNotPresentException":         {},
	"UnsupportedOperationException":   {},
	"AbstractMethodError":             {},
	"AssertionError":                  {},
	"BootstrapMethodError":            {},
	"ClassCircularityError":           {},
	"ClassFormatError":                {},
	"ExceptionInInitializerError":     {},
	"IncompatibleClassChangeError":    {},
	"InternalError":                   {},
	"LinkageError":                    {},
	"NoClassDefFoundError":            {},
	"NoSuchFieldError":                {},
	"NoSuchMethodError":               {},
	"OutOfMemoryError":                {},
	"StackOverflowError":              {},
	"UnknownError":                    {},
	"UnsatisfiedLinkError":            {},
	"UnsupportedClassVersionError":    {},
	"VerifyError":                     {},
	"VirtualMachineError":             {},
	"Deprecated":                      {},
	"FunctionalInterface":             {},
	"Override":                        {},
	"SafeVarargs":                     {},
	"SuppressWarnings":                {},
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

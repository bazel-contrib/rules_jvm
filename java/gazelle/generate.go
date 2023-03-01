package gazelle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/rule"
	bzl "github.com/bazelbuild/buildtools/build"
	"github.com/rs/zerolog"
)

type main struct {
	pkg       string
	className string
}

type javaFile struct {
	pathRelativeToBazelWorkspaceRoot string
	pkg                              string
}

func (jf *javaFile) ClassName() string {
	return strings.TrimSuffix(filepath.Base(jf.pathRelativeToBazelWorkspaceRoot), ".java")
}

func (jf *javaFile) FullyQualifiedClassName() string {
	className := ""
	if jf.pkg != "" {
		className += jf.pkg
		className += "."
	}
	className += jf.ClassName()
	return className
}

func javaFileLess(l, r javaFile) bool {
	return l.pathRelativeToBazelWorkspaceRoot < r.pathRelativeToBazelWorkspaceRoot
}

// GenerateRules extracts build metadata from source files in a directory.
//
// See language.GenerateRules for more information.
func (l javaLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	log := l.logger.With().Str("step", "GenerateRules").Str("rel", args.Rel).Logger()

	cfgs := args.Config.Exts[languageName].(javaconfig.Configs)
	cfg := cfgs[args.Rel]

	var res language.GenerateResult
	if !cfg.ExtensionEnabled() {
		return res
	}

	var protoRuleNames []string
	protoPackages := make(map[string]proto.Package)
	protoFileInfo := make(map[string]proto.FileInfo)
	for _, r := range args.OtherGen {
		if r.Kind() != "proto_library" {
			continue
		}
		pkg := r.PrivateAttr(proto.PackageKey).(proto.Package)
		protoPackages[r.Name()] = pkg
		for name, info := range pkg.Files {
			protoFileInfo[name] = info
		}
		protoRuleNames = append(protoRuleNames, r.Name())
	}
	sort.Strings(protoRuleNames)

	isModule := cfg.ModuleGranularity() == "module"

	for _, protoRuleName := range protoRuleNames {
		protoPackage := protoPackages[protoRuleName]

		jplName := strings.TrimSuffix(protoRuleName, "_proto") + "_java_proto"
		jglName := strings.TrimSuffix(protoRuleName, "_proto") + "_java_grpc"
		jlName := strings.TrimSuffix(protoRuleName, "_proto") + "_java_library"

		rjpl := rule.NewRule("java_proto_library", jplName)
		rjpl.SetAttr("deps", []string{":" + protoRuleName})
		res.Gen = append(res.Gen, rjpl)
		res.Imports = append(res.Imports, []string{})

		if protoPackage.HasServices {
			r := rule.NewRule("java_grpc_library", jglName)
			r.SetAttr("srcs", []string{":" + protoRuleName})
			r.SetAttr("deps", []string{":" + jplName})
			res.Gen = append(res.Gen, r)
			res.Imports = append(res.Imports, []string{})
		}

		rjl := rule.NewRule("java_library", jlName)
		rjl.SetAttr("visibility", []string{"//:__subpackages__"})
		var exports []string
		if protoPackage.HasServices {
			exports = append(exports, ":"+jglName)
		}
		rjl.SetAttr("exports", append(exports, ":"+jplName))
		log.Debug().Str("pkg", protoPackage.Options["java_package"]).Msg("adding the proto import statement")
		rjl.SetPrivateAttr(packagesKey, []string{protoPackage.Options["java_package"]})
		res.Gen = append(res.Gen, rjl)
		res.Imports = append(res.Imports, []string{})
	}

	javaFilenamesRelativeToPackage := filterStrSlice(args.RegularFiles, func(f string) bool { return filepath.Ext(f) == ".java" })

	if len(javaFilenamesRelativeToPackage) == 0 {
		if !isModule || !cfg.IsModuleRoot() {
			return res
		}
	}

	sort.Strings(javaFilenamesRelativeToPackage)

	javaPkg, err := l.parser.ParsePackage(context.Background(), &javaparser.ParsePackageRequest{
		Rel:   args.Rel,
		Files: javaFilenamesRelativeToPackage,
	})
	if err != nil {
		panic(err)
	}

	// We exclude intra-package imports to avoid self-dependencies.
	// This isn't a great heuristic for a few reasons:
	//  1. We may want to split targets with more granularity than per-package.
	//  2. "What input files did you have" isn't a great heuristic for "What classes are generated"
	//     (e.g. inner classes, annotation processor generated classes, etc).
	// But it will do for now.
	javaClassNamesFromFileNames := sorted_set.NewSortedSet([]string{})
	for _, filename := range javaFilenamesRelativeToPackage {
		javaClassNamesFromFileNames.Add(strings.TrimSuffix(filename, ".java"))
	}

	if isModule {
		if len(javaFilenamesRelativeToPackage) > 0 {
			l.javaPackageCache[args.Rel] = javaPkg
		}

		if !cfg.IsModuleRoot() {
			log.Debug().Msg("module // sub directory, returning early")
			if args.File != nil {
				// In module mode, there should be no intermediate build files.
				if err := os.RemoveAll(args.File.Path); err != nil {
					log.Fatal().Err(err).Msg("could not delete build file")
				}
			}
			return res
		}
	}

	var allMains []main

	// Files and imports for code which isn't tests, and isn't used as helpers in tests.
	productionJavaFiles := sorted_set.NewSortedSet([]string{})
	productionJavaImports := sorted_set.NewSortedSet([]string{})

	// Files and imports for actual test classes.
	testJavaFiles := sorted_set.NewSortedSetFn([]javaFile{}, javaFileLess)
	testJavaImports := sorted_set.NewSortedSet([]string{})

	// Java Test files which need to be generated separately from any others because they have explicit attribute overrides.
	separateTestJavaFiles := make(map[javaFile]map[string]bzl.Expr)

	// Files which are used by non-test classes in test java packages.
	testHelperJavaFiles := sorted_set.NewSortedSetFn([]javaFile{}, javaFileLess)

	// All java packages present in this bazel package.
	allPackageNames := sorted_set.NewSortedSet([]string{})

	if isModule {
		for mRel, mJavaPkg := range l.javaPackageCache {
			if !strings.HasPrefix(mRel, args.Rel) {
				continue
			}
			allPackageNames.Add(mJavaPkg.Name)

			if !mJavaPkg.TestPackage {
				addNonLocalImports(productionJavaImports, mJavaPkg.ImportedClasses, mJavaPkg.Name, javaClassNamesFromFileNames)
				addNonLocalImports(productionJavaImports, mJavaPkg.ImportedPackagesWithoutSpecificClasses, mJavaPkg.Name, javaClassNamesFromFileNames)
				for _, f := range mJavaPkg.Files.SortedSlice() {
					productionJavaFiles.Add(filepath.Join(mRel, f))
				}
				for _, className := range mJavaPkg.Mains.SortedSlice() {
					allMains = append(allMains, main{
						pkg:       mJavaPkg.Name,
						className: className,
					})
				}
			} else {
				addNonLocalImports(testJavaImports, mJavaPkg.ImportedClasses, mJavaPkg.Name, javaClassNamesFromFileNames)
				addNonLocalImports(testJavaImports, mJavaPkg.ImportedPackagesWithoutSpecificClasses, mJavaPkg.Name, javaClassNamesFromFileNames)
				for _, f := range mJavaPkg.Files.SortedSlice() {
					path := filepath.Join(mRel, f)
					file := javaFile{
						pathRelativeToBazelWorkspaceRoot: path,
						pkg:                              mJavaPkg.Name,
					}
					accumulateJavaFile(cfg, testJavaFiles, testHelperJavaFiles, separateTestJavaFiles, file, mJavaPkg.PerClassMetadata, log)
				}
			}
		}
	} else {
		allPackageNames.Add(javaPkg.Name)
		if javaPkg.TestPackage {
			addNonLocalImports(testJavaImports, javaPkg.ImportedClasses, javaPkg.Name, javaClassNamesFromFileNames)
			addNonLocalImports(testJavaImports, javaPkg.ImportedPackagesWithoutSpecificClasses, javaPkg.Name, javaClassNamesFromFileNames)
		} else {
			addNonLocalImports(productionJavaImports, javaPkg.ImportedClasses, javaPkg.Name, javaClassNamesFromFileNames)
			addNonLocalImports(productionJavaImports, javaPkg.ImportedPackagesWithoutSpecificClasses, javaPkg.Name, javaClassNamesFromFileNames)
		}
		for _, mainClassName := range javaPkg.Mains.SortedSlice() {
			allMains = append(allMains, main{
				pkg:       javaPkg.Name,
				className: mainClassName,
			})
		}
		for _, f := range javaFilenamesRelativeToPackage {
			path := filepath.Join(args.Rel, f)
			if javaPkg.TestPackage {
				file := javaFile{
					pathRelativeToBazelWorkspaceRoot: path,
					pkg:                              javaPkg.Name,
				}
				accumulateJavaFile(cfg, testJavaFiles, testHelperJavaFiles, separateTestJavaFiles, file, javaPkg.PerClassMetadata, log)
			} else {
				productionJavaFiles.Add(path)
			}
		}
	}

	allPackageNamesSlice := allPackageNames.SortedSlice()
	nonLocalProductionJavaImports := productionJavaImports.Filter(func(i string) bool {
		for _, n := range allPackageNamesSlice {
			if strings.HasPrefix(i, n) {
				// Assume the standard java convention of class names starting with upper case
				// and package components starting with lower case.
				// Without this check, one module with dependencies on a subpackage which _isn't_
				// in the module won't be detected.
				suffixRunes := []rune(i[len(n):])
				if len(suffixRunes) >= 2 && suffixRunes[0] == '.' && unicode.IsUpper(suffixRunes[1]) {
					return false
				}
			}
		}
		return true
	})

	if productionJavaFiles.Len() > 0 {
		l.generateJavaLibrary(args.File, args.Rel, filepath.Base(args.Rel), productionJavaFiles.SortedSlice(), allPackageNamesSlice, nonLocalProductionJavaImports.SortedSlice(), false, &res)
	}

	for _, m := range allMains {
		l.generateJavaBinary(args.File, m, filepath.Base(args.Rel), &res)
	}

	// We add special packages to point to testonly libraries which - this accumulates them,
	// as well as the existing java imports of tests.
	testJavaImportsWithHelpers := testJavaImports.Clone()

	if testHelperJavaFiles.Len() > 0 {
		// Suites generate their own helper library.
		if cfg.TestMode() == "file" {
			srcs := make([]string, 0, testHelperJavaFiles.Len())
			packages := sorted_set.NewSortedSet([]string{})

			for _, tf := range testHelperJavaFiles.SortedSlice() {
				// Add a _TESTONLY! prefix to the package name so that we resolve to the test-helper library rather than the production library, if both are present.
				testonlyPackage := "_TESTONLY!" + tf.pkg
				packages.Add(testonlyPackage)
				testJavaImportsWithHelpers.Add(testonlyPackage)
				srcs = append(srcs, tf.pathRelativeToBazelWorkspaceRoot)
			}
			l.generateJavaLibrary(args.File, args.Rel, filepath.Base(args.Rel), srcs, packages.SortedSlice(), testJavaImports.SortedSlice(), true, &res)
		}
	}

	allTestRelatedSrcs := testJavaFiles.Clone()
	allTestRelatedSrcs.AddAll(testHelperJavaFiles)

	if allTestRelatedSrcs.Len() > 0 {
		switch cfg.TestMode() {
		case "file":
			for _, tf := range testJavaFiles.SortedSlice() {
				extraAttributes := separateTestJavaFiles[tf]
				l.generateJavaTest(args.File, args.Rel, tf, isModule, testJavaImportsWithHelpers, extraAttributes, &res)
			}

		case "suite":
			packageNames := sorted_set.NewSortedSet([]string{})
			for _, tf := range allTestRelatedSrcs.SortedSlice() {
				packageNames.Add(tf.pkg)
			}

			suiteName := filepath.Base(args.Rel)
			if isModule {
				suiteName += "-tests"
			}

			srcs := make([]string, 0, allTestRelatedSrcs.Len())
			for _, src := range allTestRelatedSrcs.SortedSlice() {
				if _, ok := separateTestJavaFiles[src]; !ok {
					srcs = append(srcs, strings.TrimPrefix(src.pathRelativeToBazelWorkspaceRoot, args.Rel+"/"))
				}
			}
			if len(srcs) > 0 {
				l.generateJavaTestSuite(
					args.File,
					suiteName,
					srcs,
					packageNames,
					testJavaImportsWithHelpers,
					cfg.GetCustomJavaTestFileSuffixes(),
					&res,
				)
			}

			sortedSeparateTestJavaFiles := sorted_set.NewSortedSetFn([]javaFile{}, javaFileLess)
			for src := range separateTestJavaFiles {
				sortedSeparateTestJavaFiles.Add(src)
			}
			for _, src := range sortedSeparateTestJavaFiles.SortedSlice() {
				l.generateJavaTest(args.File, args.Rel, src, isModule, testJavaImportsWithHelpers, separateTestJavaFiles[src], &res)
			}
		}
	}

	for i := 0; i < len(res.Gen); i++ {
		log.Debug().Fields(map[string]interface{}{
			"idx":     i,
			"rule":    fmt.Sprintf("%#v", res.Gen[i]),
			"imports": fmt.Sprintf("%#v", res.Imports[i]),
		}).Msg("generate return")
	}

	return res
}

func (l javaLang) collectRuntimeDeps(kind, name string, file *rule.File) *sorted_set.SortedSet[label.Label] {
	runtimeDeps := sorted_set.NewSortedSetFn([]label.Label{}, labelLess)
	if file == nil {
		return runtimeDeps
	}

	for _, r := range file.Rules {
		if r.Kind() != kind || r.Name() != name {
			continue
		}

		// This does not support non string list values from runtime_deps.
		// Currently, that means if a target has a runtime_deps of a different
		// kind (e.g. a select), we will remove it. Hopefully in the future we
		// can be less destructive.
		for _, dep := range r.AttrStrings("runtime_deps") {
			parsedLabel, err := label.Parse(dep)
			if err != nil {
				l.logger.Fatal().
					Str("file.Pkg", file.Pkg).
					Str("name", name).
					Str("dep", dep).
					Err(err).
					Msg("label parse error")
			}
			runtimeDeps.Add(parsedLabel)
		}
		break
	}

	return runtimeDeps
}

// We exclude intra-target imports because otherwise we'd get self-dependencies come resolve time.
func addNonLocalImports(to *sorted_set.SortedSet[string], from *sorted_set.SortedSet[string], pkg string, localClasses *sorted_set.SortedSet[string]) {
	for _, impString := range from.SortedSlice() {
		imp := java.NewImport(impString)
		if pkg == imp.Pkg {
			if localClasses.Contains(imp.Classes[0]) {
				continue
			}
		}
		if imp.Pkg == "" {
			continue
		}

		to.Add(impString)
	}
}

func accumulateJavaFile(cfg *javaconfig.Config, testJavaFiles, testHelperJavaFiles *sorted_set.SortedSet[javaFile], separateTestJavaFiles map[javaFile]map[string]bzl.Expr, file javaFile, perClassMetadata map[string]java.PerClassMetadata, log zerolog.Logger) {
	if cfg.IsJavaTestFile(filepath.Base(file.pathRelativeToBazelWorkspaceRoot)) {
		annotationClassNames := perClassMetadata[file.FullyQualifiedClassName()].AnnotationClassNames
		perFileAttrs := make(map[string]bzl.Expr)
		for _, annotationClassName := range annotationClassNames.SortedSlice() {
			if attrs, ok := cfg.AttributesForAnnotation(annotationClassName); ok {
				for k, v := range attrs {
					if old, ok := perFileAttrs[k]; ok {
						log.Error().Str("file", file.pathRelativeToBazelWorkspaceRoot).Msgf("Saw conflicting attr overrides from annotations for attribute %v: %v and %v. Picking one at random.", k, old, v)
					}
					perFileAttrs[k] = v
				}
			}
		}
		testJavaFiles.Add(file)
		if len(perFileAttrs) > 0 {
			separateTestJavaFiles[file] = perFileAttrs
		}
	} else {
		testHelperJavaFiles.Add(file)
	}
}

func (l javaLang) generateJavaLibrary(file *rule.File, pathToPackageRelativeToBazelWorkspace string, name string, srcsRelativeToBazelWorkspace []string, packages []string, imports []string, testonly bool, res *language.GenerateResult) {
	const ruleKind = "java_library"
	r := rule.NewRule(ruleKind, name)

	srcs := make([]string, 0, len(srcsRelativeToBazelWorkspace))
	for _, src := range srcsRelativeToBazelWorkspace {
		srcs = append(srcs, strings.TrimPrefix(src, pathToPackageRelativeToBazelWorkspace+"/"))
	}
	sort.Strings(srcs)

	runtimeDeps := l.collectRuntimeDeps(ruleKind, name, file)
	if runtimeDeps.Len() > 0 {
		r.SetAttr("runtime_deps", labelsToStrings(runtimeDeps.SortedSlice()))
	}

	r.SetAttr("srcs", srcs)
	if testonly {
		r.SetAttr("testonly", true)
	} else {
		r.SetAttr("visibility", []string{"//:__subpackages__"})
	}
	r.SetPrivateAttr(packagesKey, packages)
	res.Gen = append(res.Gen, r)
	res.Imports = append(res.Imports, imports)
}

func (l javaLang) generateJavaBinary(file *rule.File, m main, libName string, res *language.GenerateResult) {
	const ruleKind = "java_binary"
	name := m.className
	r := rule.NewRule(ruleKind, name) // FIXME check collision on name
	r.SetAttr("main_class", m.pkg+"."+m.className)

	runtimeDeps := l.collectRuntimeDeps(ruleKind, name, file)
	runtimeDeps.Add(label.Label{Name: libName, Relative: true})
	r.SetAttr("runtime_deps", labelsToStrings(runtimeDeps.SortedSlice()))

	r.SetAttr("visibility", []string{"//visibility:public"})
	res.Gen = append(res.Gen, r)
	res.Imports = append(res.Imports, []string{})
}

func (l javaLang) generateJavaTest(file *rule.File, pathToPackageRelativeToBazelWorkspace string, f javaFile, includePackageInName bool, imports *sorted_set.SortedSet[string], extraAttributes map[string]bzl.Expr, res *language.GenerateResult) {
	fullyQualifiedTestClass := f.FullyQualifiedClassName()
	var testName string
	if includePackageInName {
		testName = strings.ReplaceAll(fullyQualifiedTestClass, ".", "_")
	} else {
		testName = f.ClassName()
	}

	ruleKind := "java_test"
	if importsJunit5(imports) {
		ruleKind = "java_junit5_test"
	}

	runtimeDeps := l.collectRuntimeDeps(ruleKind, testName, file)
	if importsJunit5(imports) {
		// This should probably register imports here, and then allow the
		// resolver to resolve this to an artifact, but we don't currently wire
		// up the resolver to do this. We probably should.
		// In the mean time, hard-code some labels.
		for _, artifact := range junit5RuntimeDeps {
			runtimeDeps.Add(maven.LabelFromArtifact(artifact))
		}
	}

	r := rule.NewRule(ruleKind, testName)
	path := strings.TrimPrefix(f.pathRelativeToBazelWorkspaceRoot, pathToPackageRelativeToBazelWorkspace+"/")
	r.SetAttr("srcs", []string{path})
	r.SetAttr("test_class", fullyQualifiedTestClass)
	r.SetPrivateAttr(packagesKey, []string{f.pkg})

	if runtimeDeps.Len() != 0 {
		r.SetAttr("runtime_deps", labelsToStrings(runtimeDeps.SortedSlice()))
	}

	for k, v := range extraAttributes {
		r.SetAttr(k, v)
	}

	res.Gen = append(res.Gen, r)
	testImports := imports.Clone()
	testImports.Add(f.pkg)
	res.Imports = append(res.Imports, testImports.SortedSlice())
}

func importsJunit4(imports *sorted_set.SortedSet[string]) bool {
	return imports.Contains("org.junit.Test") || imports.Contains("org.junit")
}

// Determines whether the given import is part of the JUnit Pioneer extension pack for JUnit 5. Only the beginning of
// the string is considered here to cover classes imported from different sub-packages: org.junitpioneer.vintage.Test,
// org.junitpioneer.jupiter.RetryingTest, org.junitpioneer.jupiter.cartesian.CartesianTest, etc.
func importsJunitPioneer(import_ string) bool {
	return strings.HasPrefix(import_, "org.junitpioneer.")
}

func importsJunit5(imports *sorted_set.SortedSet[string]) bool {
	return imports.Contains("org.junit.jupiter.api.Test") ||
		imports.Contains("org.junit.jupiter.api") ||
		imports.Contains("org.junit.jupiter.params.ParameterizedTest") ||
		imports.Contains("org.junit.jupiter.params") ||
		imports.Filter(importsJunitPioneer).Len() != 0
}

var junit5RuntimeDeps = []string{
	"org.junit.jupiter:junit-jupiter-engine",
	"org.junit.platform:junit-platform-launcher",
	"org.junit.platform:junit-platform-reporting",
}

func (l javaLang) generateJavaTestSuite(file *rule.File, name string, srcs []string, packageNames, imports *sorted_set.SortedSet[string], customTestSuffixes *[]string, res *language.GenerateResult) {
	const ruleKind = "java_test_suite"
	r := rule.NewRule(ruleKind, name)
	r.SetAttr("srcs", srcs)
	r.SetPrivateAttr(packagesKey, packageNames.SortedSlice())

	runtimeDeps := l.collectRuntimeDeps(ruleKind, name, file)
	if importsJunit5(imports) {
		r.SetAttr("runner", "junit5")
		for _, artifact := range junit5RuntimeDeps {
			runtimeDeps.Add(maven.LabelFromArtifact(artifact))
		}
		if importsJunit4(imports) {
			runtimeDeps.Add(maven.LabelFromArtifact("org.junit.vintage:junit-vintage-engine"))
		}
		// This should probably register imports here, and then allow the resolver to resolve this to an artifact,
		// but we don't currently wire up the resolver to do this.
		// We probably should.
		// In the mean time, hard-code some labels.
		r.SetAttr("runtime_deps", labelsToStrings(runtimeDeps.SortedSlice()))
	}

	if customTestSuffixes != nil {
		r.SetAttr("test_suffixes", *customTestSuffixes)
	}

	res.Gen = append(res.Gen, r)
	suiteImports := imports.Clone()
	suiteImports.AddAll(packageNames)
	res.Imports = append(res.Imports, suiteImports.SortedSlice())
}

func filterStrSlice(elts []string, f func(string) bool) []string {
	var out []string
	for _, elt := range elts {
		if !f(elt) {
			continue
		}
		out = append(out, elt)
	}
	return out
}

func labelsToStrings(labels []label.Label) []string {
	out := make([]string, len(labels))
	for idx, l := range labels {
		out[idx] = l.String()
	}
	return out
}

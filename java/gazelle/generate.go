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
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

type main struct {
	pkg       string
	className string
}

type javaFile struct {
	pathRelativeToBazelWorkspaceRoot string
	pkg                              string
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
			productionJavaImports.AddAll(mJavaPkg.Imports)

			if !mJavaPkg.TestPackage {
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
				for _, f := range mJavaPkg.Files.SortedSlice() {
					path := filepath.Join(mRel, f)
					file := javaFile{
						pathRelativeToBazelWorkspaceRoot: path,
						pkg:                              mJavaPkg.Name,
					}
					testJavaImports.AddAll(mJavaPkg.Imports)
					if maven.IsTestFile(filepath.Base(f)) {
						testJavaFiles.Add(file)
					} else {
						testHelperJavaFiles.Add(file)
					}
				}
			}
		}
	} else {
		allPackageNames.Add(javaPkg.Name)
		if javaPkg.TestPackage {
			testJavaImports.AddAll(javaPkg.Imports)
		} else {
			productionJavaImports.AddAll(javaPkg.Imports)
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
				if maven.IsTestFile(filepath.Base(f)) {
					testJavaFiles.Add(file)
				} else {
					testHelperJavaFiles.Add(file)
				}
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

	for _, m := range allMains {
		generateJavaBinary(m, filepath.Base(args.Rel), &res)
	}

	if productionJavaFiles.Len() > 0 {
		generateJavaLibrary(args.Rel, filepath.Base(args.Rel), productionJavaFiles.SortedSlice(), allPackageNamesSlice, nonLocalProductionJavaImports.SortedSlice(), false, &res)
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
			generateJavaLibrary(args.Rel, filepath.Base(args.Rel), srcs, packages.SortedSlice(), testJavaImports.SortedSlice(), true, &res)
		}
	}

	allTestRelatedSrcs := testJavaFiles.Clone()
	allTestRelatedSrcs.AddAll(testHelperJavaFiles)

	if allTestRelatedSrcs.Len() > 0 {
		switch cfg.TestMode() {
		case "file":
			for _, tf := range testJavaFiles.SortedSlice() {
				generateJavaTest(args.Rel, tf, isModule, testJavaImportsWithHelpers, &res)
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
				srcs = append(srcs, strings.TrimPrefix(src.pathRelativeToBazelWorkspaceRoot, args.Rel+"/"))
			}
			generateJavaTestSuite(
				suiteName,
				srcs,
				packageNames,
				testJavaImportsWithHelpers,
				&res,
			)
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

func generateJavaLibrary(pathToPackageRelativeToBazelWorkspace string, name string, srcsRelativeToBazelWorkspace []string, packages []string, imports []string, testonly bool, res *language.GenerateResult) {
	r := rule.NewRule("java_library", name)

	srcs := make([]string, 0, len(srcsRelativeToBazelWorkspace))
	for _, src := range srcsRelativeToBazelWorkspace {
		srcs = append(srcs, strings.TrimPrefix(src, pathToPackageRelativeToBazelWorkspace+"/"))
	}
	sort.Strings(srcs)

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

func generateJavaBinary(m main, libName string, res *language.GenerateResult) {
	r := rule.NewRule("java_binary", m.className) // FIXME check collision on name
	r.SetAttr("main_class", m.pkg+"."+m.className)
	r.SetAttr("runtime_deps", []string{":" + libName})
	r.SetAttr("visibility", []string{"//visibility:public"})
	res.Gen = append(res.Gen, r)
	res.Imports = append(res.Imports, []string{})
}

func generateJavaTest(pathToPackageRelativeToBazelWorkspace string, f javaFile, includePackageInName bool, imports *sorted_set.SortedSet[string], res *language.GenerateResult) {
	className := strings.TrimSuffix(filepath.Base(f.pathRelativeToBazelWorkspaceRoot), ".java")
	fullyQualifiedTestClass := f.pkg + "." + className
	var testName string
	if includePackageInName {
		testName = strings.ReplaceAll(fullyQualifiedTestClass, ".", "_")
	} else {
		testName = className
	}
	r := rule.NewRule("java_test", testName)
	path := strings.TrimPrefix(f.pathRelativeToBazelWorkspaceRoot, pathToPackageRelativeToBazelWorkspace+"/")
	r.SetAttr("srcs", []string{path})
	r.SetAttr("test_class", fullyQualifiedTestClass)
	r.SetPrivateAttr(packagesKey, []string{f.pkg})

	res.Gen = append(res.Gen, r)
	testImports := imports.Clone()
	testImports.Add(f.pkg)
	res.Imports = append(res.Imports, testImports.SortedSlice())
}

var junit5RuntimeDeps = []string{
	"org.junit.jupiter:junit-jupiter-engine",
	"org.junit.platform:junit-platform-launcher",
	"org.junit.platform:junit-platform-reporting",
}

func generateJavaTestSuite(name string, srcs []string, packageNames, imports *sorted_set.SortedSet[string], res *language.GenerateResult) {
	r := rule.NewRule("java_test_suite", name)
	r.SetAttr("srcs", srcs)
	r.SetPrivateAttr(packagesKey, packageNames.SortedSlice())

	r.SetAttr("runner", "junit5")
	r.SetAttr("runtime_deps", mapStringSlice(junit5RuntimeDeps, maven.LabelFromArtifact))

	res.Gen = append(res.Gen, r)
	suiteImports := imports.Clone()
	suiteImports.AddAll(packageNames)
	res.Imports = append(res.Imports, suiteImports.SortedSlice())
}

func strSliceUniq(elts []string) []string {
	m := make(map[string]bool, len(elts))
	for _, elt := range elts {
		m[elt] = true
	}
	var out []string
	for elt := range m {
		out = append(out, elt)
	}
	sort.Strings(out)
	return out
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

func mapStringSlice(elts []string, f func(string) string) []string {
	var out []string
	for _, elt := range elts {
		out = append(out, f(elt))
	}
	return out
}

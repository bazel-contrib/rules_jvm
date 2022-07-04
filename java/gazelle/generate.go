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
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

type main struct {
	pkg       string
	className string
}

type javaFile struct {
	path string
	pkg  string
}

type javaFiles []javaFile

func (x javaFiles) Len() int           { return len(x) }
func (x javaFiles) Less(i, j int) bool { return x[i].path < x[j].path }
func (x javaFiles) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

// GenerateRules extracts build metadata from source files in a directory.
//
// See language.GenerateRules for more information.
func (l javaLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	log := l.logger.With().Str("step", "GenerateRules").Str("rel", args.Rel).Logger()

	cfgs := args.Config.Exts[languageName].(javaconfig.Configs)
	cfg := cfgs[args.Rel]

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

	var res language.GenerateResult
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

	javaFilenames := filterStrSlice(args.RegularFiles, func(f string) bool { return filepath.Ext(f) == ".java" })
	if len(javaFilenames) == 0 {
		if isModule && cfg.IsModuleRoot() {
			l.generateModuleRoot(args, cfg, &res)
		}

		return res
	}
	sort.Strings(javaFilenames)

	javaPkg, err := l.parser.ParsePackage(context.Background(), &javaparser.ParsePackageRequest{
		Rel:   args.Rel,
		Files: javaFilenames,
	})
	if err != nil {
		panic(err)
	}

	if isModule {
		if len(javaFilenames) > 0 {
			l.javaPackageCache[args.Rel] = javaPkg
		}

		if cfg.IsModuleRoot() {
			l.generateModuleRoot(args, cfg, &res)
		} else {
			log.Debug().Msg("module // sub directory, returning early")
			if args.File != nil {
				// In module mode, there should be no intermediate build files.
				if err := os.RemoveAll(args.File.Path); err != nil {
					log.Fatal().Err(err).Msg("could not delete build file")
				}
			}
		}

		return res
	}

	if java.IsTestPath(args.Rel) {
		var javaFiles javaFiles
		for _, f := range javaFilenames {
			javaFiles = append(javaFiles, javaFile{
				path: f,
				pkg:  javaPkg.Name,
			})
		}
		generateTestRules(args, cfg, javaFiles, javaPkg.Imports, false, &res)
	} else {
		r := rule.NewRule("java_library", filepath.Base(args.Rel))
		r.SetAttr("srcs", javaFilenames)
		r.SetAttr("visibility", []string{"//:__subpackages__"})
		r.SetPrivateAttr(packagesKey, []string{javaPkg.Name})
		res.Gen = append(res.Gen, r)
		imports := javaPkg.Imports
		sort.Strings(imports)
		res.Imports = append(res.Imports, imports)
	}

	for _, m := range javaPkg.Mains {
		r := rule.NewRule("java_binary", m)
		r.SetAttr("main_class", javaPkg.Name+"."+m)
		r.SetAttr("runtime_deps", []string{":" + filepath.Base(args.Rel)})
		r.SetAttr("visibility", []string{"//visibility:public"})

		res.Gen = append(res.Gen, r)
		res.Imports = append(res.Imports, []string{})
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

func (l javaLang) generateModuleRoot(args language.GenerateArgs, cfg *javaconfig.Config, res *language.GenerateResult) {
	log := l.logger.With().Str("step", "generateModuleRoot").Str("rel", args.Rel).Logger()

	log.Debug().
		Str("granularity", cfg.ModuleGranularity()).
		Bool("isModuleRoot", cfg.IsModuleRoot()).
		Msg("GenerateRules")

	var allPackageNames []string
	allImports := make(map[string]bool)
	var allTestImports []string
	var allMains []main
	var allJavaFilenames []string
	var allTestJavaFilenames javaFiles

	for mRel, javaPkg := range l.javaPackageCache {
		if !strings.HasPrefix(mRel, args.Rel) {
			continue
		}

		log.Debug().
			Str("rel", mRel).
			Str("package", javaPkg.Name).
			Strs("imports", javaPkg.Imports).
			Strs("mains", javaPkg.Mains).
			Msg("java package cache")

		allPackageNames = append(allPackageNames, javaPkg.Name)

		if !javaPkg.TestPackage {
			for _, imp := range javaPkg.Imports {
				allImports[imp] = true
			}
			for _, m := range javaPkg.Mains {
				allMains = append(allMains, main{
					pkg:       javaPkg.Name,
					className: m,
				})
			}
			for _, f := range javaPkg.Files {
				allJavaFilenames = append(allJavaFilenames, filepath.Join(mRel, f))
			}
		} else {
			allTestImports = append(allTestImports, javaPkg.Imports...)
			for _, f := range javaPkg.Files {
				allTestJavaFilenames = append(allTestJavaFilenames, javaFile{
					path: filepath.Join(mRel, f),
					pkg:  javaPkg.Name,
				})
			}
		}
	}

	sort.Strings(allJavaFilenames)
	allPackageNames = strSliceUniq(allPackageNames)
	sort.Strings(allPackageNames)
	sort.Sort(allTestJavaFilenames)

	filteredImports := filterImports(allImports, func(i string) bool {
		for _, n := range allPackageNames {
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

	var srcs []string
	for _, f := range allJavaFilenames {
		srcs = append(srcs, strings.TrimPrefix(f, args.Rel+"/"))
	}

	r := rule.NewRule("java_library", filepath.Base(args.Rel))
	r.SetAttr("srcs", srcs)
	r.SetAttr("visibility", []string{"//:__subpackages__"})
	r.SetPrivateAttr(packagesKey, allPackageNames)
	res.Gen = append(res.Gen, r)
	sort.Strings(filteredImports)
	res.Imports = append(res.Imports, filteredImports)

	for _, m := range allMains {
		r := rule.NewRule("java_binary", m.className) // FIXME check collision on name
		r.SetAttr("main_class", m.pkg+"."+m.className)
		r.SetAttr("runtime_deps", []string{":" + filepath.Base(args.Rel)})
		res.Gen = append(res.Gen, r)
		res.Imports = append(res.Imports, []string{})
	}

	generateTestRules(args, cfg, allTestJavaFilenames, allTestImports, true, res)
}

func generateTestRules(args language.GenerateArgs, cfg *javaconfig.Config, javaFilenames javaFiles, imports []string, moduleRoot bool, res *language.GenerateResult) {
	switch cfg.TestMode() {
	case "file":
		if moduleRoot {
			for _, tf := range javaFilenames {
				if !maven.IsTestFile(filepath.Base(tf.path)) {
					continue
				}

				testClass := tf.pkg + "." + strings.TrimSuffix(filepath.Base(tf.path), ".java")
				testName := strings.ReplaceAll(testClass, ".", "_")
				r := rule.NewRule("java_test", testName)
				r.SetAttr("srcs", []string{strings.TrimPrefix(tf.path, args.Rel+"/")})
				r.SetAttr("test_class", testClass)
				r.SetPrivateAttr(packagesKey, []string{tf.pkg})

				res.Gen = append(res.Gen, r)
				imports := append(imports, tf.pkg)
				sort.Strings(imports)
				res.Imports = append(res.Imports, imports)
			}
		} else {
			var testHelperFiles []string
			for _, f := range javaFilenames {
				if maven.IsTestFile(filepath.Base(f.path)) {
					testHelperFiles = append(testHelperFiles, f.path)
				}
			}

			for _, f := range javaFilenames {
				if !maven.IsTestFile(filepath.Base(f.path)) {
					continue
				}
				makeSingleJavaTest(f, testHelperFiles, imports, res)
			}
		}

	case "suite":
		hasTest := false
		for _, f := range javaFilenames {
			if maven.IsTestFile(f.path) {
				hasTest = true
				break
			}
		}

		packageNames := make(map[string]bool)
		for _, f := range javaFilenames {
			packageNames[f.pkg] = true
		}
		var packageNamesStrSlice []string
		for pkg := range packageNames {
			packageNamesStrSlice = append(packageNamesStrSlice, pkg)
		}
		sort.Strings(packageNamesStrSlice)

		if moduleRoot {
			if hasTest {
				var srcs []string
				for _, tf := range javaFilenames {
					srcs = append(srcs, strings.TrimPrefix(tf.path, args.Rel+"/"))
				}

				javaTestSuite(
					filepath.Base(args.Rel)+"-tests",
					srcs,
					packageNamesStrSlice,
					imports,
					res,
				)
			}
		} else {
			if hasTest {
				var srcs []string
				for _, tf := range javaFilenames {
					srcs = append(srcs, strings.TrimPrefix(tf.path, args.Rel+"/"))
				}

				javaTestSuite(
					filepath.Base(args.Rel),
					srcs,
					packageNamesStrSlice,
					imports,
					res,
				)
			} else {
				r := rule.NewRule("java_library", filepath.Base(args.Rel))
				var srcs []string
				for _, tf := range javaFilenames {
					srcs = append(srcs, strings.TrimPrefix(tf.path, args.Rel+"/"))
				}
				r.SetAttr("srcs", srcs)
				r.SetAttr("testonly", true)
				r.SetPrivateAttr(packagesKey, packageNamesStrSlice)
				res.Gen = append(res.Gen, r)
				imports := append(imports, packageNamesStrSlice...)
				sort.Strings(imports)
				res.Imports = append(res.Imports, imports)
			}
		}
	}
}

func makeSingleJavaTest(f javaFile, testHelperFiles []string, imports []string, res *language.GenerateResult) {
	testName := strings.TrimSuffix(f.path, ".java")
	r := rule.NewRule("java_test", testName)
	r.SetAttr("srcs", []string{f.path})
	r.SetAttr("test_class", f.pkg+"."+testName)
	r.SetPrivateAttr(packagesKey, []string{f.pkg})
	if len(testHelperFiles) > 0 {
		r.SetAttr("deps", testHelperFiles)
	}

	res.Gen = append(res.Gen, r)
	res.Imports = append(res.Imports, append(imports, f.pkg))
}

var junit5RuntimeDeps = []string{
	"org.junit.jupiter:junit-jupiter-engine",
	"org.junit.platform:junit-platform-launcher",
	"org.junit.platform:junit-platform-reporting",
}

func javaTestSuite(name string, srcs, packageNames, imports []string, res *language.GenerateResult) {
	r := rule.NewRule("java_test_suite", name)
	r.SetAttr("srcs", srcs)
	r.SetPrivateAttr(packagesKey, packageNames)

	r.SetAttr("runner", "junit5")
	r.SetAttr("runtime_deps", mapStringSlice(junit5RuntimeDeps, maven.LabelFromArtifact))

	res.Gen = append(res.Gen, r)
	imports = append(imports, packageNames...)
	sort.Strings(imports)
	res.Imports = append(res.Imports, imports)
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

func filterImports(imports map[string]bool, f func(imp string) bool) []string {
	var out []string
	for imp := range imports {
		if f(imp) {
			out = append(out, imp)
		}
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

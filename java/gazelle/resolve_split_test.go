package gazelle

import (
	"path/filepath"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// TestSplitPackageTestSuiteHelperResolution covers a test helper whose package
// has multiple production providers. Package resolution is ambiguous in that
// case, so class-level resolution must still consider the helper library emitted
// by java_test_suite.
func TestSplitPackageTestSuiteHelperResolution(t *testing.T) {
	c, langs, _ := testConfig(t)
	mrslv, exts := InitTestResolversAndExtensions(langs)
	ix := resolve.NewRuleIndex(mrslv.Resolver, exts...)
	rc := testRemoteCache(nil)

	var jLang *javaLang
	for _, lang := range langs {
		if jl, ok := lang.(*javaLang); ok {
			jLang = jl
			break
		}
	}
	if jLang == nil {
		t.Fatal("javaLang not found in test config")
	}

	javaPackage := types.NewPackageName("com.example.shared")
	productionContent := `java_library(
    name = "one",
    _packages = ["com.example.shared"],
)

java_library(
    name = "two",
    _packages = ["com.example.shared"],
)
`
	productionFile, err := rule.LoadData(filepath.Join("production", "BUILD.bazel"), "production", []byte(productionContent))
	if err != nil {
		t.Fatal(err)
	}
	for i, r := range productionFile.Rules {
		setPackagesPrivateAttr(r)
		providerLabel := label.New("", "production", r.Name())
		jLang.classExportCache[providerLabel.String()] = classExportInfo{
			classes:  []types.ClassName{types.NewClassName(javaPackage, []string{"One", "Two"}[i])},
			testonly: false,
		}
		ix.AddRule(c, r, productionFile)
	}

	suiteContent := `java_test_suite(
    name = "suite",
)
`
	suiteFile, err := rule.LoadData(filepath.Join("helpers", "BUILD.bazel"), "helpers", []byte(suiteContent))
	if err != nil {
		t.Fatal(err)
	}
	suiteRule := suiteFile.Rules[0]
	suiteRule.SetPrivateAttr(packagesKey, []types.ResolvableJavaPackage{
		*types.NewResolvableJavaPackage(javaPackage, true, true),
	})
	helperLabel := label.New("", "helpers", "suite-test-lib")
	jLang.classExportCache[helperLabel.String()] = classExportInfo{
		classes:  []types.ClassName{types.NewClassName(javaPackage, "Helper")},
		testonly: true,
	}
	ix.AddRule(c, suiteRule, suiteFile)
	ix.Finish()

	productionSpec := resolve.ImportSpec{Lang: languageName, Imp: types.NewResolvableJavaPackage(javaPackage, false, false).String()}
	if got := len(ix.FindRulesByImportWithConfig(c, productionSpec, languageName)); got != 2 {
		t.Fatalf("test precondition violated: got %d production providers, want 2", got)
	}

	importerContent := `java_test_suite(
    name = "consumer",
)
`
	importerFile, err := rule.LoadData("BUILD.bazel", "", []byte(importerContent))
	if err != nil {
		t.Fatal(err)
	}
	importerRule := importerFile.Rules[0]
	resolveInput := types.ResolveInput{
		PackageNames:         testPackageNames(),
		ImportedPackageNames: testPackageNames(javaPackage),
		ImportedClasses:      testClassNames(types.NewClassName(javaPackage, "Helper")),
		ExportedPackageNames: testPackageNames(),
		ExportedClassNames:   testClassNames(),
		AnnotationProcessors: testClassNames(),
	}

	mrslv.Resolver(importerRule, "").Resolve(c, ix, rc, importerRule, resolveInput, label.New("", "", "consumer"))

	got := importerRule.AttrStrings("deps")
	if len(got) != 1 || got[0] != "//helpers:suite-test-lib" {
		t.Errorf("deps mismatch: got %v, want [//helpers:suite-test-lib]", got)
	}
}

func testPackageNames(values ...types.PackageName) *sorted_set.SortedSet[types.PackageName] {
	return sorted_set.NewSortedSetFn(values, types.PackageNameLess)
}

func testClassNames(values ...types.ClassName) *sorted_set.SortedSet[types.ClassName] {
	return sorted_set.NewSortedSetFn(values, types.ClassNameLess)
}

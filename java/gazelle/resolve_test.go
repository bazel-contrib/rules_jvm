package gazelle

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/language/proto"
	"github.com/bazelbuild/bazel-gazelle/pathtools"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/bazel-gazelle/testtools"
	"github.com/bazelbuild/bazel-gazelle/walk"
	bzl "github.com/bazelbuild/buildtools/build"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/tools/go/vcs"
)

func TestImports(t *testing.T) {
	type buildFile struct {
		rel, content string
	}

	type wantImport struct {
		importSpec resolve.ImportSpec
		labelName  string
	}

	type testCase struct {
		old  buildFile
		want []wantImport
	}

	for name, tc := range map[string]testCase{
		"java": {
			old: buildFile{
				rel: "",
				content: `load("@rules_java//java:defs.bzl", "java_library")

java_library(
    name = "hello",
    srcs = ["Hello.java"],
	_packages = ["com.example"],
    visibility = ["//:__subpackages__"],
)`,
			},
			want: []wantImport{
				wantImport{
					importSpec: resolve.ImportSpec{
						Lang: "java",
						Imp:  "com.example",
					},
					labelName: "hello",
				},
			},
		},
		"kotlin": {
			old: buildFile{
				rel: "",
				content: `load("@rules_kotlin//kotlin:jvm.bzl", "kt_jvm_library")

# gazelle:jvm_kotlin_enabled true

kt_jvm_library(
    name = "hello",
    srcs = ["Hello.kt"],
	_packages = ["com.example"],
    visibility = ["//:__subpackages__"],
)`,
			},
			want: []wantImport{
				wantImport{
					importSpec: resolve.ImportSpec{
						Lang: "java",
						Imp:  "com.example",
					},
					labelName: "hello",
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			c, langs, confs := testConfig(t)

			mrslv, exts := InitTestResolversAndExtensions(langs)
			ix := resolve.NewRuleIndex(mrslv.Resolver, exts...)

			buildPath := filepath.Join(filepath.FromSlash(tc.old.rel), "BUILD.bazel")
			f, err := rule.LoadData(buildPath, tc.old.rel, []byte(tc.old.content))
			if err != nil {
				t.Fatal(err)
			}
			for _, configurer := range confs {
				// Update the config to handle gazelle directives in the BUILD file.
				configurer.Configure(c, tc.old.rel, f)
			}
			for _, r := range f.Rules {
				// Explicitly set the private `_java_packages` attribute for import resolution,
				// This must be done manually as all the attributes stated in the BUILD file are
				// considered public.
				setPackagesPrivateAttr(r)
				ix.AddRule(c, r, f)
				t.Logf("added rule %s", r.Name())
			}
			ix.Finish()

			for _, want := range tc.want {
				results := ix.FindRulesByImportWithConfig(c, want.importSpec, "java")
				if len(results) != 1 {
					t.Errorf("expected 1 result, got %d for import %v", len(results), want.importSpec.Imp)
				} else {
					if results[0].Label.Name != want.labelName {
						t.Errorf("expected label %s, got %s", want.labelName, results[0].Label)
					}
				}
			}
		})
	}
}

func TestResolve(t *testing.T) {
	type buildFile struct {
		rel, content string
	}

	type testCase struct {
		old  buildFile
		want string
	}

	for name, tc := range map[string]testCase{
		"internal": {
			old: buildFile{
				rel: "",
				content: `load("@rules_java//java:defs.bzl", "java_library")

java_library(
    name = "hello",
    srcs = ["Hello.java"],
    _imported_packages = ["java.lang"],
    _packages = ["com.example"],
    visibility = ["//:__subpackages__"],
)`,
			},
			want: `load("@rules_java//java:defs.bzl", "java_library")

java_library(
    name = "hello",
    srcs = ["Hello.java"],
    visibility = ["//:__subpackages__"],
)`,
		},
		"external": {
			old: buildFile{
				rel: "",
				content: `load("@rules_java//java:defs.bzl", "java_library")

java_library(
    name = "myproject",
    srcs = ["App.java"],
    _imported_packages = [
        "com.google.common.primitives",
        "java.lang",
    ],
    _packages = ["com.example"],
    visibility = ["//:__subpackages__"],
)`,
			},
			want: `load("@rules_java//java:defs.bzl", "java_library")

java_library(
    name = "myproject",
    srcs = ["App.java"],
    visibility = ["//:__subpackages__"],
    deps = ["@maven//:com_google_guava_guava"],
)`,
		},
		"kotlin": {
			old: buildFile{
				rel: "",
				content: `load("@rules_kotlin//kotlin:jvm.bzl", "kt_jvm_library")

# gazelle:jvm_kotlin_enabled true

kt_jvm_library(
    name = "myproject",
    srcs = ["App.kt"],
    _imported_packages = [
        "com.google.common.primitives",
        "kotlin.collections",
    ],
    _packages = ["com.example"],
    visibility = ["//:__subpackages__"],
)`,
			},
			want: `load("@rules_kotlin//kotlin:jvm.bzl", "kt_jvm_library")

# gazelle:jvm_kotlin_enabled true

kt_jvm_library(
    name = "myproject",
    srcs = ["App.kt"],
    visibility = ["//:__subpackages__"],
    deps = ["@maven//:com_google_guava_guava"],
)`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			c, langs, _ := testConfig(t)

			mrslv, exts := InitTestResolversAndExtensions(langs)
			ix := resolve.NewRuleIndex(mrslv.Resolver, exts...)
			rc := testRemoteCache(nil)

			buildPath := filepath.Join(filepath.FromSlash(tc.old.rel), "BUILD.bazel")
			f, err := rule.LoadData(buildPath, tc.old.rel, []byte(tc.old.content))
			if err != nil {
				t.Fatal(err)
			}
			imports := make([]interface{}, len(f.Rules))
			for i, r := range f.Rules {
				imports[i] = convertImportsAttr(r)
				ix.AddRule(c, r, f)
			}
			ix.Finish()
			for i, r := range f.Rules {
				mrslv.Resolver(r, "").Resolve(c, ix, rc, r, imports[i], label.New("", tc.old.rel, r.Name()))

				if r.Attr("deps") != nil {
					for _, dep := range r.Attr("deps").(*bzl.ListExpr).List {
						if _, ok := dep.(*bzl.StringExpr); !ok {
							t.Errorf("rule %s deps deps should have type []StringExpr", r.Name())
						}
					}
				}
			}
			f.Sync()
			got := strings.TrimSpace(string(bzl.Format(f.File)))
			want := strings.TrimSpace(tc.want)
			if got != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(want, got, true)
				t.Errorf("Resolve:\n%s", dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func testRemoteCache(knownRepos []repo.Repo) *repo.RemoteCache {
	rc, _ := repo.NewRemoteCache(knownRepos)
	rc.RepoRootForImportPath = stubRepoRootForImportPath
	rc.HeadCmd = func(_, _ string) (string, error) {
		return "", fmt.Errorf("HeadCmd not supported in test")
	}
	rc.ModInfo = stubModInfo
	return rc
}

// stubRepoRootForImportPath is a stub implementation of vcs.RepoRootForImportPath
func stubRepoRootForImportPath(importPath string, verbose bool) (*vcs.RepoRoot, error) {
	if pathtools.HasPrefix(importPath, "example.com/repo.git") {
		return &vcs.RepoRoot{
			VCS:  vcs.ByCmd("git"),
			Repo: "https://example.com/repo.git",
			Root: "example.com/repo.git",
		}, nil
	}

	if pathtools.HasPrefix(importPath, "example.com/repo") {
		return &vcs.RepoRoot{
			VCS:  vcs.ByCmd("git"),
			Repo: "https://example.com/repo.git",
			Root: "example.com/repo",
		}, nil
	}

	if pathtools.HasPrefix(importPath, "example.com") {
		return &vcs.RepoRoot{
			VCS:  vcs.ByCmd("git"),
			Repo: "https://example.com",
			Root: "example.com",
		}, nil
	}

	return nil, fmt.Errorf("could not resolve import path: %q", importPath)
}

// stubModInfo is a stub implementation of RemoteCache.ModInfo.
func stubModInfo(importPath string) (string, error) {
	if pathtools.HasPrefix(importPath, "example.com/repo/v2") {
		return "example.com/repo/v2", nil
	}
	if pathtools.HasPrefix(importPath, "example.com/repo") {
		return "example.com/repo", nil
	}
	return "", fmt.Errorf("could not find module for import path: %q", importPath)
}

func setPackagesPrivateAttr(r *rule.Rule) {
	packages := r.AttrStrings("_packages")
	resolvablePackages := make([]types.ResolvableJavaPackage, 0, len(packages))
	for _, pkg := range packages {
		pkgName := types.NewPackageName(pkg)
		resolvablePackages = append(resolvablePackages, *types.NewResolvableJavaPackage(pkgName, false, false))
	}
	r.SetPrivateAttr(packagesKey, resolvablePackages)
}

func convertImportsAttr(r *rule.Rule) types.ResolveInput {
	return types.ResolveInput{
		PackageNames:         packageAttrToSortedSet(r, "_packages"),
		ImportedPackageNames: packageAttrToSortedSet(r, "_imported_packages"),
	}
}

func packageAttrToSortedSet(r *rule.Rule, name string) *sorted_set.SortedSet[types.PackageName] {
	attrValues := r.AttrStrings(name)
	r.DelAttr(name)
	packages := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	for _, v := range attrValues {
		packages.Add(types.NewPackageName(v))
	}
	return packages
}

func testConfig(t *testing.T, args ...string) (*config.Config, []language.Language, []config.Configurer) {
	// Add a -repo_root argument if none is present. Without this,
	// config.CommonConfigurer will try to auto-detect a WORKSPACE file,
	// which will fail.
	haveRoot := false
	for _, arg := range args {
		if strings.HasPrefix(arg, "-repo_root") {
			haveRoot = true
			break
		}
	}
	if !haveRoot {
		args = append(args, "-repo_root=.")
	}

	cexts := []config.Configurer{
		new(config.CommonConfigurer),
		new(walk.Configurer),
		new(resolve.Configurer),
	}

	l := NewLanguage()
	l.(*javaLang).mavenResolver = &testResolver{}

	langs := []language.Language{
		proto.NewLanguage(),
		l,
	}

	c := testtools.NewTestConfig(t, cexts, langs, args)

	absRepoRoot, err := filepath.Abs(c.RepoRoot)
	if err != nil {
		t.Fatalf("error getting absolute path for workspace")
	}
	c.RepoRoot = absRepoRoot

	for _, lang := range langs {
		cexts = append(cexts, lang)
	}

	return c, langs, cexts
}

type testResolver struct{}

func (*testResolver) Resolve(pkg types.PackageName, excludedArtifacts map[string]struct{}, mavenRepositoryName string) (label.Label, error) {
	return label.NoLabel, errors.New("not implemented")
}

func (*testResolver) ResolveClass(className types.ClassName, excludedArtifacts map[string]struct{}, mavenRepositoryName string) (label.Label, error) {
	return label.NoLabel, errors.New("not implemented")
}

type mapResolver map[string]resolve.Resolver

func (mr mapResolver) Resolver(r *rule.Rule, f string) resolve.Resolver {
	return mr[r.Kind()]
}

func InitTestResolversAndExtensions(langs []language.Language) (mapResolver, []interface{}) {
	mrslv := make(mapResolver)
	exts := make([]interface{}, 0, len(langs))
	for _, lang := range langs {
		// TODO There has to be a better way to make this generic.
		if jLang, ok := lang.(*javaLang); ok {
			jLang.mavenResolver = NewTestMavenResolver()
			jLang.javaExportIndex.FinalizeIndex()
		}

		for kind := range lang.Kinds() {
			mrslv[kind] = lang
		}
		exts = append(exts, lang)
	}
	return mrslv, exts
}

type TestMavenResolver struct {
	data map[types.PackageName]label.Label
}

func NewTestMavenResolver() *TestMavenResolver {
	return &TestMavenResolver{
		data: map[types.PackageName]label.Label{
			types.NewPackageName("com.google.common.primitives"): label.New("maven", "", "com_google_guava_guava"),
			types.NewPackageName("org.junit"):                    label.New("maven", "", "junit_junit"),
		},
	}
}

func (r *TestMavenResolver) Resolve(pkg types.PackageName, excludedArtifacts map[string]struct{}, mavenRepositoryName string) (label.Label, error) {
	l, found := r.data[pkg]
	if !found {
		return label.NoLabel, fmt.Errorf("unexpected import: %s", pkg)
	}
	return l, nil
}

func (r *TestMavenResolver) ResolveClass(className types.ClassName, excludedArtifacts map[string]struct{}, mavenRepositoryName string) (label.Label, error) {
	return label.NoLabel, nil
}

func TestProtoSplitPackageClassResolution(t *testing.T) {
	c, langs, _ := testConfig(t)

	mrslv, exts := InitTestResolversAndExtensions(langs)
	ix := resolve.NewRuleIndex(mrslv.Resolver, exts...)

	pkg := "protos/logging"
	javaPackage := types.NewPackageName("com.example.protos.logging.http")

	httpLibContent := `load("@rules_java//java:defs.bzl", "java_library")

java_library(
    name = "http_java_library",
    _packages = ["com.example.protos.logging.http"],
    exports = [":http_java_proto"],
)
`
	sawmillLibContent := `load("@rules_java//java:defs.bzl", "java_library")

java_library(
    name = "sawmill_raw_http_request_java_library",
    _packages = ["com.example.protos.logging.http"],
    exports = [":sawmill_raw_http_request_java_proto"],
)
`

	buildPath := filepath.Join(filepath.FromSlash(pkg), "BUILD.bazel")

	httpFile, err := rule.LoadData(buildPath, pkg, []byte(httpLibContent))
	if err != nil {
		t.Fatal(err)
	}

	sawmillFile, err := rule.LoadData(buildPath, pkg, []byte(sawmillLibContent))
	if err != nil {
		t.Fatal(err)
	}

	var javaLangInstance *javaLang
	for _, lang := range langs {
		if jl, ok := lang.(*javaLang); ok {
			javaLangInstance = jl
			break
		}
	}
	if javaLangInstance == nil {
		t.Fatal("javaLang not found in langs")
	}

	httpRule := httpFile.Rules[0]
	setPackagesPrivateAttr(httpRule)
	httpLabel := label.New("", pkg, "http_java_library")
	javaLangInstance.classExportCache[httpLabel.String()] = classExportInfo{
		classes: []types.ClassName{
			types.NewClassName(javaPackage, "Http"),
			types.NewClassName(javaPackage, "Request"),
		},
		testonly: false,
	}
	ix.AddRule(c, httpRule, httpFile)

	sawmillRule := sawmillFile.Rules[0]
	setPackagesPrivateAttr(sawmillRule)
	sawmillLabel := label.New("", pkg, "sawmill_raw_http_request_java_library")
	javaLangInstance.classExportCache[sawmillLabel.String()] = classExportInfo{
		classes: []types.ClassName{
			types.NewClassName(javaPackage, "SawmillRawHttpRequest"),
		},
		testonly: false,
	}
	ix.AddRule(c, sawmillRule, sawmillFile)

	ix.Finish()

	importSpec := resolve.ImportSpec{Lang: "java", Imp: javaPackage.Name}
	matches := ix.FindRulesByImportWithConfig(c, importSpec, "java")
	if len(matches) != 2 {
		t.Fatalf("expected 2 providers for the package, got %d", len(matches))
	}

	resolver := NewResolver(javaLangInstance)

	pci := resolver.buildPackageClassIndex(c, javaPackage, ix)

	if _, ok := pci.prod["Http"]; !ok {
		t.Error("Http class should be indexed")
	} else if len(pci.prod["Http"]) != 1 {
		t.Errorf("Http should have exactly 1 provider, got %d", len(pci.prod["Http"]))
	} else if pci.prod["Http"][0] != httpLabel {
		t.Errorf("Http should be provided by http_java_library, got %s", pci.prod["Http"][0])
	}

	if _, ok := pci.prod["Request"]; !ok {
		t.Error("Request class should be indexed")
	} else if len(pci.prod["Request"]) != 1 {
		t.Errorf("Request should have exactly 1 provider, got %d", len(pci.prod["Request"]))
	} else if pci.prod["Request"][0] != httpLabel {
		t.Errorf("Request should be provided by http_java_library, got %s", pci.prod["Request"][0])
	}

	if _, ok := pci.prod["SawmillRawHttpRequest"]; !ok {
		t.Error("SawmillRawHttpRequest class should be indexed")
	} else if len(pci.prod["SawmillRawHttpRequest"]) != 1 {
		t.Errorf("SawmillRawHttpRequest should have exactly 1 provider, got %d", len(pci.prod["SawmillRawHttpRequest"]))
	} else if pci.prod["SawmillRawHttpRequest"][0] != sawmillLabel {
		t.Errorf("SawmillRawHttpRequest should be provided by sawmill_raw_http_request_java_library, got %s", pci.prod["SawmillRawHttpRequest"][0])
	}
}

// fakeExternalPluginResolver simulates an external plugin (like rules_wire)
// that returns Java ImportSpecs for its custom rule kinds.
type fakeExternalPluginResolver struct {
	// javaPackages maps rule names to the Java packages they provide
	javaPackages map[string][]types.ResolvableJavaPackage
}

func (r *fakeExternalPluginResolver) Name() string {
	return "java" // Returns "java" so ImportSpecs are indexed under the Java language
}

func (r *fakeExternalPluginResolver) Imports(c *config.Config, rule *rule.Rule, f *rule.File) []resolve.ImportSpec {
	pkgs, ok := r.javaPackages[rule.Name()]
	if !ok {
		return nil
	}
	var specs []resolve.ImportSpec
	for _, pkg := range pkgs {
		specs = append(specs, resolve.ImportSpec{Lang: "java", Imp: pkg.String()})
	}
	return specs
}

func (r *fakeExternalPluginResolver) Embeds(rule *rule.Rule, from label.Label) []label.Label {
	return nil
}

func (r *fakeExternalPluginResolver) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, rule *rule.Rule, imports interface{}, from label.Label) {
}

func TestSharedClassCacheForExternalPlugins(t *testing.T) {
	c, langs, _ := testConfig(t)

	mrslv, exts := InitTestResolversAndExtensions(langs)

	var javaLangInstance *javaLang
	for _, lang := range langs {
		if jl, ok := lang.(*javaLang); ok {
			javaLangInstance = jl
			break
		}
	}
	if javaLangInstance == nil {
		t.Fatal("javaLang not found")
	}

	pkg := "splitpkg"
	javaPackage := types.NewPackageName("com.example.split")

	// Create a fake external plugin resolver for java_wire_library.
	// In production, this would be the Wire plugin returning Java ImportSpecs.
	fakeWireResolver := &fakeExternalPluginResolver{
		javaPackages: map[string][]types.ResolvableJavaPackage{
			"wire_proto": {*types.NewResolvableJavaPackage(javaPackage, false, false)},
		},
	}
	mrslv["java_wire_library"] = fakeWireResolver

	ix := resolve.NewRuleIndex(mrslv.Resolver, exts...)

	// Create a java_library that provides the package (represents hand-written code)
	javaLibContent := `
java_library(
    name = "java_part",
    srcs = ["JavaPart.java"],
)
`
	buildPath := filepath.Join(filepath.FromSlash(pkg), "BUILD.bazel")
	javaFile, err := rule.LoadData(buildPath, pkg, []byte(javaLibContent))
	if err != nil {
		t.Fatal(err)
	}
	javaRule := javaFile.Rules[0]

	// Set up the java_library with its package
	javaRule.SetPrivateAttr(packagesKey, []types.ResolvableJavaPackage{
		*types.NewResolvableJavaPackage(javaPackage, false, false),
	})

	// Set up classExportCache for java_library (simulating generation)
	javaLabel := label.New("", pkg, "java_part")
	javaLangInstance.classExportCache[javaLabel.String()] = classExportInfo{
		classes:  []types.ClassName{types.NewClassName(javaPackage, "JavaPart")},
		testonly: false,
	}

	// Create a java_wire_library rule (external plugin's rule kind)
	wireLibContent := `
java_wire_library(
    name = "wire_proto",
    proto = ":person_proto",
)
`
	wireFile, err := rule.LoadData(buildPath, pkg, []byte(wireLibContent))
	if err != nil {
		t.Fatal(err)
	}
	wireRule := wireFile.Rules[0]

	// External plugin contributes class info via shared cache (not classExportCache).
	// This is the key difference: internal rules use classExportCache,
	// external plugins use SharedClassCache.
	sharedCache := javaconfig.GetOrCreateSharedClassCache(c)
	wireLabel := label.New("", pkg, "wire_proto")
	sharedCache[wireLabel.String()] = javaconfig.SharedClassInfo{
		Classes:  []string{"com.example.split.WireMessage"},
		TestOnly: false,
	}

	// Add both rules to the index
	ix.AddRule(c, javaRule, javaFile)
	ix.AddRule(c, wireRule, wireFile)
	ix.Finish()

	resolver := NewResolver(javaLangInstance)

	// Build the class index for this split package
	pci := resolver.buildPackageClassIndex(c, javaPackage, ix)

	// The java_library class should be indexed from classExportCache
	if _, ok := pci.prod["JavaPart"]; !ok {
		t.Error("JavaPart should be indexed from java_library's classExportCache")
	}

	// The java_wire_library class should be indexed from SharedClassCache
	if _, ok := pci.prod["WireMessage"]; !ok {
		t.Error("WireMessage should be indexed from SharedClassCache")
	}

	// Verify both classes map to their respective labels
	if len(pci.prod["JavaPart"]) != 1 || pci.prod["JavaPart"][0].Name != "java_part" {
		t.Errorf("JavaPart should map to java_part, got %v", pci.prod["JavaPart"])
	}
	if len(pci.prod["WireMessage"]) != 1 || pci.prod["WireMessage"][0].Name != "wire_proto" {
		t.Errorf("WireMessage should map to wire_proto, got %v", pci.prod["WireMessage"])
	}
}

func TestClassLevelResolutionInSplitPackage(t *testing.T) {
	c, langs, _ := testConfig(t)

	mrslv, exts := InitTestResolversAndExtensions(langs)
	ix := resolve.NewRuleIndex(mrslv.Resolver, exts...)

	var javaLangInstance *javaLang
	for _, lang := range langs {
		if jl, ok := lang.(*javaLang); ok {
			javaLangInstance = jl
			break
		}
	}
	if javaLangInstance == nil {
		t.Fatal("javaLang not found")
	}

	pkg := "splitpkg"
	javaPackage := types.NewPackageName("com.example.split")

	// Create two java_library rules in the same package - simulating a split package scenario
	javaLib1Content := `
java_library(
    name = "java_part1",
    srcs = ["JavaPart1.java"],
)
`
	javaLib2Content := `
java_library(
    name = "java_part2",
    srcs = ["JavaPart2.java"],
)
`

	buildPath := filepath.Join(filepath.FromSlash(pkg), "BUILD.bazel")

	javaFile1, err := rule.LoadData(buildPath, pkg, []byte(javaLib1Content))
	if err != nil {
		t.Fatal(err)
	}
	javaFile2, err := rule.LoadData(buildPath, pkg, []byte(javaLib2Content))
	if err != nil {
		t.Fatal(err)
	}

	javaRule1 := javaFile1.Rules[0]
	javaRule2 := javaFile2.Rules[0]

	// Set up both rules with the same package (split package scenario)
	javaRule1.SetPrivateAttr(packagesKey, []types.ResolvableJavaPackage{
		*types.NewResolvableJavaPackage(javaPackage, false, false),
	})
	javaRule2.SetPrivateAttr(packagesKey, []types.ResolvableJavaPackage{
		*types.NewResolvableJavaPackage(javaPackage, false, false),
	})

	// Set up classExportCache for both rules (simulating what happens during generation)
	javaLabel1 := label.New("", pkg, "java_part1")
	javaLangInstance.classExportCache[javaLabel1.String()] = classExportInfo{
		classes:  []types.ClassName{types.NewClassName(javaPackage, "ClassA")},
		testonly: false,
	}

	javaLabel2 := label.New("", pkg, "java_part2")
	javaLangInstance.classExportCache[javaLabel2.String()] = classExportInfo{
		classes:  []types.ClassName{types.NewClassName(javaPackage, "ClassB")},
		testonly: false,
	}

	// Add rules to the index
	ix.AddRule(c, javaRule1, javaFile1)
	ix.AddRule(c, javaRule2, javaFile2)
	ix.Finish()

	// Verify both rules are indexed for the package (split package)
	importSpec := resolve.ImportSpec{Lang: "java", Imp: javaPackage.Name}
	matches := ix.FindRulesByImportWithConfig(c, importSpec, "java")
	if len(matches) != 2 {
		t.Fatalf("expected 2 providers for the package, got %d", len(matches))
	}

	resolver := NewResolver(javaLangInstance)

	// Build the class index for this split package
	pci := resolver.buildPackageClassIndex(c, javaPackage, ix)

	// Both classes should be indexed
	if _, ok := pci.prod["ClassA"]; !ok {
		t.Error("ClassA should be indexed from java_part1")
	}
	if _, ok := pci.prod["ClassB"]; !ok {
		t.Error("ClassB should be indexed from java_part2")
	}

	// Verify correct labels
	if len(pci.prod["ClassA"]) != 1 || pci.prod["ClassA"][0] != javaLabel1 {
		t.Errorf("ClassA should be provided by java_part1, got %v", pci.prod["ClassA"])
	}
	if len(pci.prod["ClassB"]) != 1 || pci.prod["ClassB"][0] != javaLabel2 {
		t.Errorf("ClassB should be provided by java_part2, got %v", pci.prod["ClassB"])
	}
}

func TestSharedClassCacheTestOnlyHandling(t *testing.T) {
	c, langs, _ := testConfig(t)

	mrslv, exts := InitTestResolversAndExtensions(langs)
	ix := resolve.NewRuleIndex(mrslv.Resolver, exts...)

	var javaLangInstance *javaLang
	for _, lang := range langs {
		if jl, ok := lang.(*javaLang); ok {
			javaLangInstance = jl
			break
		}
	}
	if javaLangInstance == nil {
		t.Fatal("javaLang not found")
	}

	pkg := "testpkg"
	javaPackage := types.NewPackageName("com.example.test")

	// Create a testonly java_library (simulates test utilities from external plugin)
	testLibContent := `
java_library(
    name = "test_helper",
    srcs = ["TestHelper.java"],
    testonly = True,
)
`
	buildPath := filepath.Join(filepath.FromSlash(pkg), "BUILD.bazel")
	testFile, err := rule.LoadData(buildPath, pkg, []byte(testLibContent))
	if err != nil {
		t.Fatal(err)
	}
	testRule := testFile.Rules[0]

	// Set up the testonly rule with its package (using testonly=true ImportSpec)
	testRule.SetPrivateAttr(packagesKey, []types.ResolvableJavaPackage{
		*types.NewResolvableJavaPackage(javaPackage, true, false), // testonly=true
	})

	// External plugin contributes testonly class info via shared cache
	sharedCache := javaconfig.GetOrCreateSharedClassCache(c)
	testLabel := label.New("", pkg, "test_helper")
	sharedCache[testLabel.String()] = javaconfig.SharedClassInfo{
		Classes:  []string{"com.example.test.TestHelper"},
		TestOnly: true,
	}

	// Create a prod java_library
	prodLibContent := `
java_library(
    name = "prod_helper",
    srcs = ["ProdHelper.java"],
)
`
	prodFile, err := rule.LoadData(buildPath, pkg, []byte(prodLibContent))
	if err != nil {
		t.Fatal(err)
	}
	prodRule := prodFile.Rules[0]

	// Set up the prod rule with its package
	prodRule.SetPrivateAttr(packagesKey, []types.ResolvableJavaPackage{
		*types.NewResolvableJavaPackage(javaPackage, false, false),
	})

	// External plugin contributes prod class info via shared cache
	prodLabel := label.New("", pkg, "prod_helper")
	sharedCache[prodLabel.String()] = javaconfig.SharedClassInfo{
		Classes:  []string{"com.example.test.ProdHelper"},
		TestOnly: false,
	}

	// Add rules to the index
	ix.AddRule(c, testRule, testFile)
	ix.AddRule(c, prodRule, prodFile)
	ix.Finish()

	resolver := NewResolver(javaLangInstance)

	// Build the class index for this package
	pci := resolver.buildPackageClassIndex(c, javaPackage, ix)

	// ProdHelper should be in prod index
	if _, ok := pci.prod["ProdHelper"]; !ok {
		t.Error("ProdHelper should be in prod index")
	}

	// TestHelper should be in test index (not prod)
	if _, ok := pci.test["TestHelper"]; !ok {
		t.Error("TestHelper should be in test index")
	}
	if _, ok := pci.prod["TestHelper"]; ok {
		t.Error("TestHelper should NOT be in prod index")
	}
}

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

func TestIsJvmLibraryWithExtensionKinds(t *testing.T) {
	c := &config.Config{
		Exts: make(map[string]interface{}),
	}

	// Built-in kinds should be recognized
	if !isJvmLibrary(c, "java_library") {
		t.Error("java_library should be recognized as JVM library")
	}
	if !isJvmLibrary(c, "kt_jvm_library") {
		t.Error("kt_jvm_library should be recognized as JVM library")
	}
	if !isJvmLibrary(c, "java_proto_library") {
		t.Error("java_proto_library should be recognized as JVM library")
	}
	if !isJvmLibrary(c, "java_grpc_library") {
		t.Error("java_grpc_library should be recognized as JVM library")
	}

	// Proto library helper functions should work correctly
	if !isJavaProtoLibrary(c, "java_proto_library") {
		t.Error("java_proto_library should be recognized as Java proto library")
	}
	if !isJavaProtoLibrary(c, "java_grpc_library") {
		t.Error("java_grpc_library should be recognized as Java proto library")
	}
	if isJavaProtoLibrary(c, "java_library") {
		t.Error("java_library should not be recognized as Java proto library")
	}

	// Unknown kinds should not be recognized by default
	if isJvmLibrary(c, "custom_jvm_library") {
		t.Error("custom_jvm_library should not be recognized without registration")
	}

	// Register a custom kind via the extension mechanism
	extKinds := make(map[string]bool)
	extKinds["custom_jvm_library"] = true
	extKinds["java_wire_library"] = true
	c.Exts[javaconfig.JavaExtensionLibraryKindsKey] = extKinds

	// Now custom kinds should be recognized
	if !isJvmLibrary(c, "custom_jvm_library") {
		t.Error("custom_jvm_library should be recognized after registration")
	}
	if !isJvmLibrary(c, "java_wire_library") {
		t.Error("java_wire_library should be recognized after registration")
	}

	// Unregistered kinds should still not be recognized
	if isJvmLibrary(c, "some_other_library") {
		t.Error("some_other_library should not be recognized")
	}

	// Built-in kinds should still work
	if !isJvmLibrary(c, "java_library") {
		t.Error("java_library should still be recognized")
	}
}

func TestExternalPluginContributedPackages(t *testing.T) {
	c, langs, _ := testConfig(t)

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

	// Register java_wire_library as a JVM library kind
	extKinds := make(map[string]bool)
	extKinds["java_wire_library"] = true
	c.Exts[javaconfig.JavaExtensionLibraryKindsKey] = extKinds

	pkg := "protos"
	wireLibContent := `java_wire_library(
    name = "person_wire",
    proto = ":person_proto",
)
`
	wireFile, err := rule.LoadData("BUILD.bazel", pkg, []byte(wireLibContent))
	if err != nil {
		t.Fatalf("failed to load BUILD file: %v", err)
	}

	wireRule := wireFile.Rules[0]
	wireRule.SetPrivateAttr(javaconfig.JavaGazelleProvidedPackagesAttr, []string{"com.example.proto"})

	resolver := NewResolver(javaLangInstance)
	imports := resolver.Imports(c, wireRule, wireFile)

	// Should have registered the package
	if len(imports) != 1 {
		t.Fatalf("expected 1 import spec, got %d", len(imports))
	}
	if imports[0].Lang != "java" {
		t.Errorf("expected lang 'java', got %s", imports[0].Lang)
	}
	if imports[0].Imp != "com.example.proto" {
		t.Errorf("expected import 'com.example.proto', got %s", imports[0].Imp)
	}
}

func TestExternalPluginContributedClasses(t *testing.T) {
	c, langs, _ := testConfig(t)

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

	pkg := "protos"
	wireLibContent := `java_wire_library(
    name = "person_wire",
    proto = ":person_proto",
)
`
	wireFile, err := rule.LoadData("BUILD.bazel", pkg, []byte(wireLibContent))
	if err != nil {
		t.Fatalf("failed to load BUILD file: %v", err)
	}

	wireRule := wireFile.Rules[0]
	// Only provide classes, not packages - should infer packages
	wireRule.SetPrivateAttr(javaconfig.JavaGazelleProvidedClassesAttr, []string{
		"com.example.proto.Person",
		"com.example.proto.Address",
	})

	resolver := NewResolver(javaLangInstance)
	imports := resolver.Imports(c, wireRule, wireFile)

	// Should have inferred and registered the package from class names
	if len(imports) != 1 {
		t.Fatalf("expected 1 import spec (inferred package), got %d", len(imports))
	}
	if imports[0].Imp != "com.example.proto" {
		t.Errorf("expected inferred package 'com.example.proto', got %s", imports[0].Imp)
	}

	// Should have populated classExportCache
	lbl := label.New("", pkg, "person_wire")
	info, ok := javaLangInstance.classExportCache[lbl.String()]
	if !ok {
		t.Fatal("classExportCache should contain the wire rule")
	}
	if len(info.classes) != 2 {
		t.Errorf("expected 2 classes in cache, got %d", len(info.classes))
	}

	// Verify the classes were parsed correctly
	classNames := make(map[string]bool)
	for _, cls := range info.classes {
		classNames[cls.BareOuterClassName()] = true
	}
	if !classNames["Person"] {
		t.Error("Person class should be in cache")
	}
	if !classNames["Address"] {
		t.Error("Address class should be in cache")
	}
}

func TestExternalPluginClassesWithExplicitPackages(t *testing.T) {
	c, langs, _ := testConfig(t)

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

	pkg := "protos"
	wireLibContent := `java_wire_library(
    name = "person_wire",
    proto = ":person_proto",
)
`
	wireFile, err := rule.LoadData("BUILD.bazel", pkg, []byte(wireLibContent))
	if err != nil {
		t.Fatalf("failed to load BUILD file: %v", err)
	}

	wireRule := wireFile.Rules[0]
	// Provide both packages and classes
	wireRule.SetPrivateAttr(javaconfig.JavaGazelleProvidedPackagesAttr, []string{"com.example.proto"})
	wireRule.SetPrivateAttr(javaconfig.JavaGazelleProvidedClassesAttr, []string{
		"com.example.proto.Person",
		"com.example.proto.Address",
	})

	resolver := NewResolver(javaLangInstance)
	imports := resolver.Imports(c, wireRule, wireFile)

	// Should have exactly 1 import (explicit package, not inferred)
	if len(imports) != 1 {
		t.Fatalf("expected 1 import spec, got %d", len(imports))
	}

	// Should still have populated classExportCache
	lbl := label.New("", pkg, "person_wire")
	info, ok := javaLangInstance.classExportCache[lbl.String()]
	if !ok {
		t.Fatal("classExportCache should contain the wire rule")
	}
	if len(info.classes) != 2 {
		t.Errorf("expected 2 classes in cache, got %d", len(info.classes))
	}
}

func TestExternalPluginClassesInSplitPackageResolution(t *testing.T) {
	c, langs, _ := testConfig(t)

	// Register java_wire_library as a JVM library kind
	extKinds := make(map[string]bool)
	extKinds["java_wire_library"] = true
	c.Exts[javaconfig.JavaExtensionLibraryKindsKey] = extKinds

	mrslv, exts := InitTestResolversAndExtensions(langs)

	// Also register java_wire_library in the resolver map so the Java resolver handles it
	// (In production, this happens because the kind is registered via JavaExtensionLibraryKindsKey)
	for _, lang := range langs {
		if _, ok := lang.(*javaLang); ok {
			mrslv["java_wire_library"] = lang
			break
		}
	}

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

	// Create two rules in the same package - simulating a split package scenario
	// One is a regular java_library, one is an external plugin rule (java_wire_library)

	javaLibContent := `
java_library(
    name = "java_part",
    srcs = ["JavaPart.java"],
    _packages = ["com.example.split"],
)
`
	wireLibContent := `
java_wire_library(
    name = "wire_part",
    proto = ":messages_proto",
)
`

	buildPath := filepath.Join(filepath.FromSlash(pkg), "BUILD.bazel")

	javaFile, err := rule.LoadData(buildPath, pkg, []byte(javaLibContent))
	if err != nil {
		t.Fatal(err)
	}
	wireFile, err := rule.LoadData(buildPath, pkg, []byte(wireLibContent))
	if err != nil {
		t.Fatal(err)
	}

	javaRule := javaFile.Rules[0]
	wireRule := wireFile.Rules[0]

	// Set up the java_library with its package (uses internal packagesKey)
	setPackagesPrivateAttr(javaRule)

	// Set up the wire rule with contributed packages and classes (simulating Wire plugin)
	// This tests the new extension mechanism where external plugins set these attrs
	wireRule.SetPrivateAttr(javaconfig.JavaGazelleProvidedPackagesAttr, []string{"com.example.split"})
	wireRule.SetPrivateAttr(javaconfig.JavaGazelleProvidedClassesAttr, []string{
		"com.example.split.WireMessage",
	})

	// Manually set up classExportCache for both rules (simulating what happens
	// when Imports() is called during the real gazelle flow)
	javaLabel := label.New("", pkg, "java_part")
	javaLangInstance.classExportCache[javaLabel.String()] = classExportInfo{
		classes:  []types.ClassName{types.NewClassName(javaPackage, "JavaPart")},
		testonly: false,
	}

	wireLabel := label.New("", pkg, "wire_part")
	javaLangInstance.classExportCache[wireLabel.String()] = classExportInfo{
		classes:  []types.ClassName{types.NewClassName(javaPackage, "WireMessage")},
		testonly: false,
	}

	// Add rules to the index - the resolver will call Imports() which reads the private attrs
	ix.AddRule(c, javaRule, javaFile)
	ix.AddRule(c, wireRule, wireFile)

	ix.Finish()

	// Verify both rules are indexed for the package
	importSpec := resolve.ImportSpec{Lang: "java", Imp: javaPackage.Name}
	matches := ix.FindRulesByImportWithConfig(c, importSpec, "java")
	if len(matches) != 2 {
		t.Fatalf("expected 2 providers for the package, got %d", len(matches))
	}

	resolver := NewResolver(javaLangInstance)

	// Build the class index for this split package
	pci := resolver.buildPackageClassIndex(c, javaPackage, ix)

	// Both classes should be indexed
	if _, ok := pci.prod["JavaPart"]; !ok {
		t.Error("JavaPart class should be indexed from java_library")
	}
	if _, ok := pci.prod["WireMessage"]; !ok {
		t.Error("WireMessage class should be indexed from java_wire_library")
	}

	// Verify correct labels
	if len(pci.prod["WireMessage"]) != 1 || pci.prod["WireMessage"][0] != wireLabel {
		t.Errorf("WireMessage should be provided by wire_part, got %v", pci.prod["WireMessage"])
	}
}

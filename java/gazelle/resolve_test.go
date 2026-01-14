package gazelle

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/maven"
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
	// Create a wrapper that implements maven.Resolver interface
	var resolver maven.Resolver = NewTestMavenResolver()
	l.(*javaLang).mavenResolvers["maven_install.json"] = &resolver

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
			// Create a wrapper that implements maven.Resolver interface
			var resolver maven.Resolver = NewTestMavenResolver()
			jLang.mavenResolvers["maven_install.json"] = &resolver
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

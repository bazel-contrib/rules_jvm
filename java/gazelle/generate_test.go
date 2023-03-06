package gazelle

import (
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/google/go-cmp/cmp"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestSingleJavaTestFile(t *testing.T) {
	f := javaFile{
		pathRelativeToBazelWorkspaceRoot: "FooTest.java",
		pkg:                              types.NewPackageName("com.example"),
	}
	type testCase struct {
		includePackageInName bool
		importedPackages     []string
		wantRuleKind         string
		wantImports          []string
		wantDeps             []string
		wantRuntimeDeps      []string
	}

	for name, tc := range map[string]testCase{
		"no imported packages no helpers no package": {
			includePackageInName: false,
			importedPackages:     nil,
			wantRuleKind:         "java_test",
			wantImports:          []string{"com.example"},
			wantDeps:             nil,
		},
		"some imported packages no helpers no package": {
			includePackageInName: false,
			importedPackages:     []string{"io.netty"},
			wantRuleKind:         "java_test",
			wantImports:          []string{"io.netty", "com.example"},
			wantDeps:             nil,
		},
		"no imported packages some helpers no package": {
			includePackageInName: false,
			importedPackages:     nil,
			wantRuleKind:         "java_test",
			wantImports:          []string{"com.example"},
			wantDeps:             []string{":helper"},
		},
		"some imported packages some helpers no package": {
			includePackageInName: false,
			importedPackages:     []string{"io.netty"},
			wantRuleKind:         "java_test",
			wantImports:          []string{"io.netty", "com.example"},
		},
		"no imported packages no helpers yes package": {
			includePackageInName: true,
			importedPackages:     nil,
			wantRuleKind:         "java_test",
			wantImports:          []string{"com.example"},
			wantDeps:             nil,
		},
		"some imported packages no helpers yes package": {
			includePackageInName: true,
			importedPackages:     []string{"io.netty"},
			wantRuleKind:         "java_test",
			wantImports:          []string{"io.netty", "com.example"},
			wantDeps:             nil,
		},
		"no imported packages some helpers yes package": {
			includePackageInName: true,
			importedPackages:     nil,
			wantRuleKind:         "java_test",
			wantImports:          []string{"com.example"},
			wantDeps:             []string{":helper"},
		},
		"some imported packages some helpers yes package": {
			includePackageInName: true,
			importedPackages:     []string{"io.netty"},
			wantRuleKind:         "java_test",
			wantImports:          []string{"io.netty", "com.example"},
		},
		"explicit junit4": {
			includePackageInName: false,
			importedPackages:     []string{"org.junit"},
			wantRuleKind:         "java_test",
			wantImports:          []string{"com.example", "org.junit"},
		},
		"explicit junit5": {
			includePackageInName: false,
			importedPackages:     []string{"org.junit.jupiter.api"},
			wantRuleKind:         "java_junit5_test",
			wantImports:          []string{"com.example", "org.junit.jupiter.api"},
			wantRuntimeDeps: []string{
				"@maven//:org_junit_jupiter_junit_jupiter_engine",
				"@maven//:org_junit_platform_junit_platform_launcher",
				"@maven//:org_junit_platform_junit_platform_reporting",
			},
		},
		"parameterized junit5": {
			includePackageInName: false,
			importedPackages:     []string{"org.junit.jupiter.params"},
			wantRuleKind:         "java_junit5_test",
			wantImports:          []string{"com.example", "org.junit.jupiter.params"},
			wantRuntimeDeps: []string{
				"@maven//:org_junit_jupiter_junit_jupiter_engine",
				"@maven//:org_junit_platform_junit_platform_launcher",
				"@maven//:org_junit_platform_junit_platform_reporting",
			},
		},
		"junitpioneer junit5": {
			includePackageInName: false,
			importedPackages:     []string{"org.junitpioneer.jupiter.cartesian"},
			wantRuleKind:         "java_junit5_test",
			wantImports:          []string{"com.example", "org.junitpioneer.jupiter.cartesian"},
			wantRuntimeDeps: []string{
				"@maven//:org_junit_jupiter_junit_jupiter_engine",
				"@maven//:org_junit_platform_junit_platform_launcher",
				"@maven//:org_junit_platform_junit_platform_reporting",
			},
		},
		"explicit both junit4 and junit5": {
			includePackageInName: false,
			importedPackages:     []string{"org.junit", "org.junit.jupiter.api"},
			wantRuleKind:         "java_junit5_test",
			wantImports:          []string{"com.example", "org.junit", "org.junit.jupiter.api"},
			wantRuntimeDeps: []string{
				"@maven//:org_junit_jupiter_junit_jupiter_engine",
				"@maven//:org_junit_platform_junit_platform_launcher",
				"@maven//:org_junit_platform_junit_platform_reporting",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var res language.GenerateResult

			l := newTestJavaLang(t)
			l.generateJavaTest(nil, "", f, tc.includePackageInName, stringsToPackageNames(tc.importedPackages), nil, &res)

			require.Len(t, res.Gen, 1, "want 1 generated rule")

			rule := res.Gen[0]
			require.Equal(t, tc.wantRuleKind, rule.Kind())
			if tc.includePackageInName {
				require.Equal(t, "com_example_FooTest", rule.AttrString("name"))
			} else {
				require.Equal(t, "FooTest", rule.AttrString("name"))
			}
			require.Equal(t, []string{"FooTest.java"}, rule.AttrStrings("srcs"))
			require.Equal(t, "com.example.FooTest", rule.AttrString("test_class"))

			wantAttrs := []string{"name", "srcs", "test_class"}
			if len(tc.wantRuntimeDeps) > 0 {
				wantAttrs = append(wantAttrs, "runtime_deps")
			}
			require.ElementsMatch(t, wantAttrs, rule.AttrKeys())

			require.Len(t, res.Imports, 1, "want 1 generated importedPackages")
			wantImports := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
			for _, wi := range tc.wantImports {
				wantImports.Add(types.NewPackageName(wi))
			}
			require.ElementsMatch(t, wantImports.SortedSlice(), res.Imports[0].(types.ResolveInput).ImportedPackageNames.SortedSlice())

			if len(tc.wantRuntimeDeps) > 0 {
				require.ElementsMatch(t, tc.wantRuntimeDeps, rule.AttrStrings("runtime_deps"))
			}
		})
	}
}

func TestSuite(t *testing.T) {
	src := "FooTest.java"
	pkg := "com.example"

	type testCase struct {
		includePackageInName bool
		importedPackages     []string
		wantImports          []string
		wantDeps             []string
		wantRuntimeDeps      []string
		wantRunner           string
	}

	for name, tc := range map[string]testCase{
		"explicit junit4": {
			includePackageInName: false,
			importedPackages:     []string{"org.junit"},
			wantImports:          []string{"com.example", "org.junit"},
		},
		"explicit junit5": {
			includePackageInName: false,
			importedPackages:     []string{"org.junit.jupiter.api"},
			wantImports:          []string{"com.example", "org.junit.jupiter.api"},
			wantRuntimeDeps: []string{
				"@maven//:org_junit_jupiter_junit_jupiter_engine",
				"@maven//:org_junit_platform_junit_platform_launcher",
				"@maven//:org_junit_platform_junit_platform_reporting",
			},
			wantRunner: "junit5",
		},
		"parameterized junit5": {
			includePackageInName: false,
			importedPackages:     []string{"org.junit.jupiter.params"},
			wantImports:          []string{"com.example", "org.junit.jupiter.params"},
			wantRuntimeDeps: []string{
				"@maven//:org_junit_jupiter_junit_jupiter_engine",
				"@maven//:org_junit_platform_junit_platform_launcher",
				"@maven//:org_junit_platform_junit_platform_reporting",
			},
			wantRunner: "junit5",
		},
		"explicit both junit4 and junit5": {
			includePackageInName: false,
			importedPackages:     []string{"org.junit", "org.junit.jupiter.api"},
			wantImports:          []string{"com.example", "org.junit", "org.junit.jupiter.api"},
			wantRuntimeDeps: []string{
				"@maven//:org_junit_jupiter_junit_jupiter_engine",
				"@maven//:org_junit_platform_junit_platform_launcher",
				"@maven//:org_junit_platform_junit_platform_reporting",
				"@maven//:org_junit_vintage_junit_vintage_engine",
			},
			wantRunner: "junit5",
		},
	} {
		t.Run(name, func(t *testing.T) {
			var res language.GenerateResult

			l := newTestJavaLang(t)
			l.generateJavaTestSuite(nil, "blah", []string{src}, stringsToPackageNames([]string{pkg}), stringsToPackageNames(tc.importedPackages), nil, false, &res)

			require.Len(t, res.Gen, 1, "want 1 generated rule")

			rule := res.Gen[0]
			require.Equal(t, "java_test_suite", rule.Kind())
			require.Equal(t, "blah", rule.AttrString("name"))
			require.Equal(t, []string{"FooTest.java"}, rule.AttrStrings("srcs"))

			wantAttrs := []string{"name", "srcs"}
			if len(tc.wantRuntimeDeps) > 0 {
				wantAttrs = append(wantAttrs, "runtime_deps")
			}
			if tc.wantRunner != "" {
				wantAttrs = append(wantAttrs, "runner")
			}
			require.ElementsMatch(t, wantAttrs, rule.AttrKeys())

			require.Len(t, res.Imports, 1, "want 1 generated importedPackages")
			wantImports := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
			for _, wi := range tc.wantImports {
				wantImports.Add(types.NewPackageName(wi))
			}
			require.ElementsMatch(t, wantImports.SortedSlice(), res.Imports[0].(types.ResolveInput).ImportedPackageNames.SortedSlice())

			if len(tc.wantRuntimeDeps) > 0 {
				require.ElementsMatch(t, tc.wantRuntimeDeps, rule.AttrStrings("runtime_deps"))
			}

			if tc.wantRunner != "" {
				require.Equal(t, tc.wantRunner, rule.AttrString("runner"))
			}
		})
	}
}

func TestAddNonLocalImports(t *testing.T) {
	src := sorted_set.NewSortedSetFn[types.ClassName]([]types.ClassName{}, types.ClassNameLess)
	for _, s := range []string{
		"com.example.a.b.Foo",        // same pkg, included class name: delete
		"com.example.a.b.Bar",        // same pkg, included class name: delete
		"com.example.a.b.Bar.SubBar", // same pkg, nested class, included class name: delete
		"com.example.a.b.Baz",        // same pkg, not included class name: keep
		"com.example.a.b.Baz.SubBaz", // same pkg, nested class, not included class name: keep
		"com.example.a.b.c.Foo",      // different pkg: keep
		"com.example.a.Foo",          // different pkg: keep
		"com.another.a.b.Foo",        // different pkg: keep
	} {
		name, err := types.ParseClassName(s)
		if err != nil {
			t.Fatal(err)
		}
		src.Add(*name)
	}

	depsDst := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	exportsDst := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	addNonLocalImportsAndExports(depsDst, exportsDst, src, sorted_set.NewSortedSetFn[types.PackageName]([]types.PackageName{}, types.PackageNameLess), sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess), types.NewPackageName("com.example.a.b"), sorted_set.NewSortedSet([]string{"Foo", "Bar"}))

	want := stringsToPackageNames([]string{
		"com.another.a.b",
		"com.example.a",
		"com.example.a.b",
		"com.example.a.b.c",
	}).SortedSlice()

	if diff := cmp.Diff(want, depsDst.SortedSlice()); diff != "" {
		t.Errorf("filterImports() mismatch (-want +got):\n%s", diff)
	}
}

func newTestJavaLang(t *testing.T) javaLang {
	t.Helper()
	return javaLang{
		logger: zerolog.New(zerolog.NewTestWriter(t)),
	}
}

func stringsToPackageNames(strs []string) *sorted_set.SortedSet[types.PackageName] {
	ret := sorted_set.NewSortedSetFn[types.PackageName]([]types.PackageName{}, types.PackageNameLess)
	for _, s := range strs {
		ret.Add(types.NewPackageName(s))
	}
	return ret
}

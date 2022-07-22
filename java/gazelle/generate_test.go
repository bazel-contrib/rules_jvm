package gazelle

import (
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/stretchr/testify/require"
)

func TestSingleJavaTestFile(t *testing.T) {
	f := javaFile{
		path: "FooTest.java",
		pkg:  "com.example",
	}
	type testCase struct {
		imports     []string
		wantImports []string

		testHelperFiles []string
		wantDeps        []string
	}

	for name, tc := range map[string]testCase{
		"no imports no helpers": {
			imports:         nil,
			wantImports:     []string{"com.example"},
			testHelperFiles: nil,
			wantDeps:        nil,
		},
		"some imports no helpers": {
			imports:         []string{"io.netty"},
			wantImports:     []string{"io.netty", "com.example"},
			testHelperFiles: nil,
			wantDeps:        nil,
		},
		"no imports some helpers": {
			imports:         nil,
			wantImports:     []string{"com.example"},
			testHelperFiles: []string{"Helper.java"},
			wantDeps:        []string{"Helper.java"},
		},
		"some imports some helpers": {
			imports:         []string{"io.netty"},
			wantImports:     []string{"io.netty", "com.example"},
			testHelperFiles: []string{"Helper.java"},
			wantDeps:        []string{"Helper.java"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var res language.GenerateResult

			makeSingleJavaTest(f, tc.testHelperFiles, sorted_set.NewSortedSet(tc.imports), &res)

			require.Len(t, res.Gen, 1, "want 1 generated rule")

			rule := res.Gen[0]
			require.Equal(t, "java_test", rule.Kind())
			require.Equal(t, "FooTest", rule.AttrString("name"))
			require.Equal(t, []string{"FooTest.java"}, rule.AttrStrings("srcs"))
			require.Equal(t, "com.example.FooTest", rule.AttrString("test_class"))
			require.Equal(t, tc.wantDeps, rule.AttrStrings("deps"))

			wantAttrs := []string{"name", "srcs", "test_class"}
			if len(tc.wantDeps) > 0 {
				wantAttrs = append(wantAttrs, "deps")
			}
			require.ElementsMatch(t, wantAttrs, rule.AttrKeys())

			require.Len(t, res.Imports, 1, "want 1 generated imports")
			require.ElementsMatch(t, tc.wantImports, res.Imports[0])
		})
	}
}

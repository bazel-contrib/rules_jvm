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

		wantDeps []string
	}

	for name, tc := range map[string]testCase{
		"no imports no helpers": {
			imports:     nil,
			wantImports: []string{"com.example"},
			wantDeps:    nil,
		},
		"some imports no helpers": {
			imports:     []string{"io.netty"},
			wantImports: []string{"io.netty", "com.example"},
			wantDeps:    nil,
		},
		"no imports some helpers": {
			imports:     nil,
			wantImports: []string{"com.example"},
			wantDeps:    []string{":helper"},
		},
		"some imports some helpers": {
			imports:     []string{"io.netty"},
			wantImports: []string{"io.netty", "com.example"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var res language.GenerateResult

			makeSingleJavaTest(f, sorted_set.NewSortedSet([]string{}), sorted_set.NewSortedSet(tc.imports), &res)

			require.Len(t, res.Gen, 1, "want 1 generated rule")

			rule := res.Gen[0]
			require.Equal(t, "java_test", rule.Kind())
			require.Equal(t, "FooTest", rule.AttrString("name"))
			require.Equal(t, []string{"FooTest.java"}, rule.AttrStrings("srcs"))
			require.Equal(t, "com.example.FooTest", rule.AttrString("test_class"))

			wantAttrs := []string{"name", "srcs", "test_class"}
			require.ElementsMatch(t, wantAttrs, rule.AttrKeys())

			require.Len(t, res.Imports, 1, "want 1 generated imports")
			require.ElementsMatch(t, tc.wantImports, res.Imports[0])
		})
	}
}

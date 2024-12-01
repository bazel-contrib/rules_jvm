package kotlin

import (
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

func TestIsStdLib(t *testing.T) {
	tests := map[string]bool{
		"":                   false,
		"kotlin":             true,
		"kotlin.math":        true,
		"kotlin.collections": true,
		"java.lang":          false,
		"com.example":        false,
	}

	for pkg, want := range tests {
		t.Run(pkg, func(t *testing.T) {
			if got := IsStdlib(types.NewPackageName(pkg)); got != want {
				t.Errorf("IsStdLib() = %v, want %v", got, want)
			}
		})
	}
}

package bazel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

func TestOutputBase(t *testing.T) {
	repoRoot, err := bazel.NewTmpDir("bar")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	f, err := os.Create(filepath.Join(repoRoot, "WORKSPACE"))
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	f.Close()

	if _, err := OutputBase(repoRoot); err != nil {
		t.Fatalf("error: %s", err)
	}
}

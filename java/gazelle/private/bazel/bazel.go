package bazel

import (
	"regexp"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
)

var (
	FindBinary   = bazel.FindBinary
	ListRunfiles = bazel.ListRunfiles
)

var nonWordRe = regexp.MustCompile(`\W+`)

func CleanupLabel(in string) string {
	return nonWordRe.ReplaceAllString(in, "_")
}

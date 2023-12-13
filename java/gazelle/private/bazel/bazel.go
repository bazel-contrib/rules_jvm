package bazel

import "regexp"

var nonWordRe = regexp.MustCompile(`\W+`)

func CleanupLabel(in string) string {
	return nonWordRe.ReplaceAllString(in, "_")
}

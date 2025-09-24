package sorted_set

import "github.com/bazelbuild/bazel-gazelle/label"

// LabelLess is a comparison function for a SortedSet that holds label.Label instances.
func LabelLess(l, r label.Label) bool {
	// In UTF-8, / sorts before :
	// We want relative labels to come before absolute ones, so explicitly sort relative before absolute.
	if l.Relative {
		if r.Relative {
			return l.String() < r.String()
		}
		return true
	}
	if r.Relative {
		return false
	}
	return l.String() < r.String()
}

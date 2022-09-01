package java

import "github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"

type Package struct {
	Name string

	Imports *sorted_set.SortedSet[string]
	Mains   *sorted_set.SortedSet[string]

	// Especially useful for module mode
	Files       *sorted_set.SortedSet[string]
	TestPackage bool

	PerClassMetadata map[string]PerClassMetadata
}

type PerClassMetadata struct {
	AnnotationClassNames *sorted_set.SortedSet[string]
}

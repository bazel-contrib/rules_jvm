package java

import (
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

type Package struct {
	Name types.PackageName

	ImportedClasses                        *sorted_set.SortedSet[types.ClassName]
	ImportedPackagesWithoutSpecificClasses *sorted_set.SortedSet[types.PackageName]
	Mains                                  *sorted_set.SortedSet[types.ClassName]

	// Especially useful for module mode
	Files       *sorted_set.SortedSet[string]
	TestPackage bool

	PerClassMetadata map[string]PerClassMetadata
}

type PerClassMetadata struct {
	AnnotationClassNames *sorted_set.SortedSet[string]
}

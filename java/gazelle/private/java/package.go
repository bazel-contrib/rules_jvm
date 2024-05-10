package java

import (
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_multiset"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

type Package struct {
	Name types.PackageName

	ImportedClasses                        *sorted_set.SortedSet[types.ClassName]
	ExportedClasses                        *sorted_set.SortedSet[types.ClassName]
	ImportedPackagesWithoutSpecificClasses *sorted_set.SortedSet[types.PackageName]
	Mains                                  *sorted_set.SortedSet[types.ClassName]

	// Especially useful for module mode
	Files       *sorted_set.SortedSet[string]
	TestPackage bool

	PerClassMetadata map[string]PerClassMetadata
}

func (p *Package) AllAnnotations() *sorted_set.SortedSet[types.ClassName] {
	annotations := sorted_set.NewSortedSetFn(nil, types.ClassNameLess)
	for _, pcm := range p.PerClassMetadata {
		annotations.AddAll(pcm.AnnotationClassNames)
		for _, method := range pcm.MethodAnnotationClassNames.Keys() {
			annotations.AddAll(pcm.MethodAnnotationClassNames.Values(method))
		}
		for _, field := range pcm.FieldAnnotationClassNames.Keys() {
			annotations.AddAll(pcm.FieldAnnotationClassNames.Values(field))
		}
	}
	return annotations
}

type PerClassMetadata struct {
	AnnotationClassNames       *sorted_set.SortedSet[types.ClassName]
	MethodAnnotationClassNames *sorted_multiset.SortedMultiSet[string, types.ClassName]
	FieldAnnotationClassNames  *sorted_multiset.SortedMultiSet[string, types.ClassName]
}

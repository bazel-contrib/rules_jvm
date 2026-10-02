package java_export_index

import (
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/rs/zerolog"
)

func TestExportIndexUsesExactClassOwnerInSplitPackage(t *testing.T) {
	index := NewJavaExportIndex("java", zerolog.Nop())
	pkg := types.NewPackageName("com.example.shared")
	first := types.NewClassName(pkg, "First")
	second := types.NewClassName(pkg, "Second")
	recordTestLibrary(index, "first", "first", pkg, first)
	recordTestLibrary(index, "second", "second", pkg, second)
	bridge := rule.NewRule("java_library", "bridge")
	index.RecordRuleWithResolveInput(rule.EmptyFile("bridge/BUILD.bazel", "bridge"), bridge, types.ResolveInput{
		ImportedPackageNames: sorted_set.NewSortedSetFn([]types.PackageName{pkg}, types.PackageNameLess),
		ImportedClasses:      sorted_set.NewSortedSetFn([]types.ClassName{second}, types.ClassNameLess),
	}, nil)
	exportFile := rule.EmptyFile("published/BUILD.bazel", "published")
	export := rule.NewRule("java_export", "published")
	export.SetAttr("exports", []string{"//bridge:bridge"})
	index.RecordJavaExport(export, exportFile)
	index.FinalizeIndex()

	if _, ok := index.IsExportedByJavaExport(label.New("", "second", "second")); !ok {
		t.Fatal("exact class owner was not included in the export")
	}
	if _, ok := index.IsExportedByJavaExport(label.New("", "first", "first")); ok {
		t.Fatal("unrelated split-package owner was included in the export")
	}
}

func TestExportIndexKeepsUniquePackageOwnerForUnindexedClass(t *testing.T) {
	index := NewJavaExportIndex("java", zerolog.Nop())
	pkg := types.NewPackageName("com.example")
	provider := label.New("", "provider", "library")
	index.RecordRuleWithResolveInput(rule.EmptyFile("provider/BUILD.bazel", "provider"), rule.NewRule("java_library", "library"), types.ResolveInput{
		PackageNames: sorted_set.NewSortedSetFn([]types.PackageName{pkg}, types.PackageNameLess),
	}, nil)
	bridge := label.New("", "bridge", "library")
	deps := index.dependenciesForResolveInput(bridge, types.ResolveInput{
		ImportedPackageNames: sorted_set.NewSortedSetFn([]types.PackageName{pkg}, types.PackageNameLess),
		ImportedClasses: sorted_set.NewSortedSetFn([]types.ClassName{
			types.NewClassName(pkg, "Unindexed"),
		}, types.ClassNameLess),
	})
	if !deps[provider] {
		t.Fatalf("unindexed class lost its unique package provider: %v", deps)
	}
}

func recordTestLibrary(index *JavaExportIndex, dir, name string, pkg types.PackageName, declared types.ClassName) {
	index.RecordRuleWithResolveInput(rule.EmptyFile(dir+"/BUILD.bazel", dir), rule.NewRule("java_library", name), types.ResolveInput{
		PackageNames: sorted_set.NewSortedSetFn([]types.PackageName{pkg}, types.PackageNameLess),
	}, sorted_set.NewSortedSetFn([]types.ClassName{declared}, types.ClassNameLess))
}

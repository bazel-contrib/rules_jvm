package scc

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

// mkPkg builds a production package: importedPkgs feed edge (a) (package imports),
// importedClasses feed both edge (a) (their package) and edge (b) (matched against
// other dirs' internalClasses), and internalClasses are the package's own internal
// symbols.
func mkPkg(name string, importedPkgs, importedClasses, internalClasses []string) *java.Package {
	pkg := &java.Package{
		Name:                                   types.NewPackageName(name),
		ImportedPackagesWithoutSpecificClasses: sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess),
		ImportedClasses:                        sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess),
		InternalClasses:                        sorted_set.NewSortedSetFn([]types.ClassName{}, types.ClassNameLess),
	}
	for _, p := range importedPkgs {
		pkg.ImportedPackagesWithoutSpecificClasses.Add(types.NewPackageName(p))
	}
	for _, c := range importedClasses {
		cn, err := types.ParseClassName(c)
		if err != nil {
			panic(err)
		}
		pkg.ImportedClasses.Add(*cn)
	}
	for _, c := range internalClasses {
		cn, err := types.ParseClassName(c)
		if err != nil {
			panic(err)
		}
		pkg.InternalClasses.Add(*cn)
	}
	return pkg
}

// groupsByName maps each group's name to its sorted member dirs, for stable comparison.
func groupsByName(g *Graph) map[string][]string {
	out := make(map[string][]string)
	for _, grp := range g.Groups() {
		out[grp.Name] = grp.Dirs
	}
	return out
}

func mustNew(t *testing.T, packages map[string]*java.Package) *Graph {
	t.Helper()
	g, err := New(packages)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return g
}

func TestNoEdgesAllSingletons(t *testing.T) {
	g := mustNew(t, map[string]*java.Package{
		"m/a": mkPkg("a", nil, nil, nil),
		"m/b": mkPkg("b", nil, nil, nil),
	})
	want := map[string][]string{"a": {"m/a"}, "b": {"m/b"}}
	if got := groupsByName(g); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestImportCycleCollapses(t *testing.T) {
	g := mustNew(t, map[string]*java.Package{
		"m/a": mkPkg("a", []string{"b"}, nil, nil),
		"m/b": mkPkg("b", []string{"a"}, nil, nil),
	})
	groups := g.Groups()
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1: %v", len(groups), groupsByName(g))
	}
	if want := []string{"m/a", "m/b"}; !reflect.DeepEqual(groups[0].Dirs, want) {
		t.Errorf("got dirs %v, want %v", groups[0].Dirs, want)
	}
}

// TestInternalEdgeCollapsesButPublicDoesNot is the crux: an acyclic reference to
// another dir's symbol collapses the two dirs iff that symbol is `internal`.
func TestInternalEdgeCollapsesButPublicDoesNot(t *testing.T) {
	// b references a.Foo, which a declares internal. No import cycle exists.
	internal := mustNew(t, map[string]*java.Package{
		"m/a": mkPkg("a", nil, nil, []string{"a.Foo"}),
		"m/b": mkPkg("b", nil, []string{"a.Foo"}, nil),
	})
	if got := len(internal.Groups()); got != 1 {
		t.Errorf("internal reference: got %d groups, want 1 (collapsed): %v", got, groupsByName(internal))
	}

	// Identical shape, but a.Foo is public (a declares no internals): stays two targets.
	public := mustNew(t, map[string]*java.Package{
		"m/a": mkPkg("a", nil, nil, nil),
		"m/b": mkPkg("b", nil, []string{"a.Foo"}, nil),
	})
	if got := len(public.Groups()); got != 2 {
		t.Errorf("public reference: got %d groups, want 2 (separate): %v", got, groupsByName(public))
	}
}

// TestInternalFunctionKeying exercises edge (b) for a top-level internal function,
// keyed as an importer records it (package + simple name, e.g. "a.fn").
func TestInternalFunctionKeying(t *testing.T) {
	g := mustNew(t, map[string]*java.Package{
		"m/a": mkPkg("a", nil, nil, []string{"a.helper"}),
		"m/b": mkPkg("b", nil, []string{"a.helper"}, nil),
	})
	if got := len(g.Groups()); got != 1 {
		t.Errorf("got %d groups, want 1 (collapsed via internal function): %v", got, groupsByName(g))
	}
}

func TestDeterministic(t *testing.T) {
	packages := map[string]*java.Package{
		"m/a": mkPkg("a", []string{"b"}, nil, nil),
		"m/b": mkPkg("b", []string{"a"}, nil, []string{"b.Secret"}),
		"m/c": mkPkg("c", []string{"b"}, []string{"b.Secret"}, nil),
		"m/d": mkPkg("d", nil, nil, nil),
	}
	first := groupsByName(mustNew(t, packages))
	// Map iteration order is randomized, so repeated runs exercise determinism.
	for i := 0; i < 10; i++ {
		if got := groupsByName(mustNew(t, packages)); !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d differs: %v vs %v", i, got, first)
		}
	}
}

func TestPackagePartition(t *testing.T) {
	packages := map[string]*java.Package{
		"m/a": mkPkg("a", []string{"b"}, nil, nil),
		"m/b": mkPkg("b", []string{"a"}, nil, nil),
		"m/c": mkPkg("c", []string{"a"}, nil, nil),
	}
	g := mustNew(t, packages)

	seen := make(map[string]int)
	for _, grp := range g.Groups() {
		for _, p := range grp.Packages.SortedSlice() {
			seen[p.Name]++
		}
	}
	want := map[string]int{"a": 1, "b": 1, "c": 1}
	if !reflect.DeepEqual(seen, want) {
		t.Errorf("package partition: got %v, want each package exactly once: %v", seen, want)
	}
	// GroupForPackage agrees with the partition.
	for name := range want {
		if g.GroupForPackage(types.NewPackageName(name)) == nil {
			t.Errorf("GroupForPackage(%q) = nil", name)
		}
	}
}

func TestNameCollisionDisambiguated(t *testing.T) {
	g := mustNew(t, map[string]*java.Package{
		"m/actions/ui":  mkPkg("actions.ui", nil, nil, nil),
		"m/eventing/ui": mkPkg("eventing.ui", nil, nil, nil),
		"m/solo":        mkPkg("solo", nil, nil, nil),
	})
	names := make([]string, 0)
	for _, grp := range g.Groups() {
		names = append(names, grp.Name)
	}
	sort.Strings(names)
	want := []string{"actions_ui", "eventing_ui", "solo"}
	if !reflect.DeepEqual(names, want) {
		t.Errorf("got names %v, want %v (collisions disambiguated, non-colliding kept as basename)", names, want)
	}
}

func TestAssertAcyclicFiresOnMisclassifiedEdge(t *testing.T) {
	// Hand-build a Graph whose grouping does NOT match the adjacency: two groups with
	// a mutual dependency. New never produces this, so we drive assertAcyclic directly.
	groupX := &Group{Dirs: []string{"m/x"}, Name: "x"}
	groupY := &Group{Dirs: []string{"m/y"}, Name: "y"}
	g := &Graph{
		groups:     []*Group{groupX, groupY},
		dirToGroup: map[string]*Group{"m/x": groupX, "m/y": groupY},
	}
	adjacency := map[string]*sorted_set.SortedSet[string]{
		"m/x": sorted_set.NewSortedSet([]string{"m/y"}),
		"m/y": sorted_set.NewSortedSet([]string{"m/x"}),
	}
	err := g.assertAcyclic(adjacency)
	if err == nil {
		t.Fatal("expected a cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error %q does not mention a cycle", err.Error())
	}
}

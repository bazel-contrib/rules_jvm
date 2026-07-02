// Package scc groups a JVM module's source directories into the minimal set of
// library targets that can compile.
//
// Bazel forbids target dependency cycles, so two directories that import each other
// cyclically must share one target. This constraint is language-agnostic -- Java
// package cycles collapse the same way Kotlin ones do. Kotlin adds a second
// constraint: one library is one Kotlin module and `internal` is module-scoped, so a
// directory referencing another's `internal` symbol must also share its target
// (rules_kotlin requires both to be the same module; an associate cannot bridge two
// modules). For a pure-Java module only the cycle constraint applies.
//
// Both constraints are expressed as edges of a directed graph over directories and
// resolved with a single strongly-connected-components pass: each SCC of size > 1
// becomes one collapsed target; every other directory stays its own target. The SCC
// condensation is a DAG, so every cross-target edge is an acyclic public import that
// Bazel can express as an ordinary `deps` edge.
package scc

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/java"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/sorted_set"
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

// Group is a set of source directories that must compile as a single library target.
type Group struct {
	// Dirs are the member directories (gazelle Rels), sorted lexicographically.
	Dirs []string
	// Packages are the packages declared by the group's directories.
	Packages *sorted_set.SortedSet[types.PackageName]
	// Name is a deterministic target name, unique among the module's groups (they all
	// share one BUILD file in module mode).
	Name string
}

// Graph is the SCC condensation of one module's directory dependency graph.
type Graph struct {
	groups         []*Group
	dirToGroup     map[string]*Group
	packageToGroup map[types.PackageName]*Group
}

// Groups returns the collapse groups, ordered by their smallest member directory.
func (g *Graph) Groups() []*Group { return g.groups }

// GroupForDir returns the group owning dir, or nil if dir is not in the module.
func (g *Graph) GroupForDir(dir string) *Group { return g.dirToGroup[dir] }

// GroupForPackage returns the group declaring pkg, or nil if pkg is not in the module.
func (g *Graph) GroupForPackage(pkg types.PackageName) *Group { return g.packageToGroup[pkg] }

// New computes the collapse for one module from its production packages, keyed by
// source directory (gazelle Rel). Directories that must share a target end up in the
// same Group; every other directory is its own singleton Group.
func New(packagesByDir map[string]*java.Package) (*Graph, error) {
	dirs := make([]string, 0, len(packagesByDir))
	for dir := range packagesByDir {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	// A package is normally declared by exactly one directory; tolerate a split
	// package by mapping to every directory that declares it.
	packageToDirs := make(map[types.PackageName][]string)
	for _, dir := range dirs {
		pkg := packagesByDir[dir]
		packageToDirs[pkg.Name] = append(packageToDirs[pkg.Name], dir)
	}

	// Fully-qualified name of an `internal` symbol -> directories declaring it.
	internalOwners := make(map[string][]string)
	for _, dir := range dirs {
		internal := packagesByDir[dir].InternalClasses
		if internal == nil {
			continue
		}
		for _, cn := range internal.SortedSlice() {
			fqn := cn.FullyQualifiedClassName()
			internalOwners[fqn] = append(internalOwners[fqn], dir)
		}
	}

	adjacency := make(map[string]*sorted_set.SortedSet[string], len(dirs))
	for _, dir := range dirs {
		adjacency[dir] = sorted_set.NewSortedSet([]string{})
	}
	addEdge := func(from, to string) {
		if from != to {
			adjacency[from].Add(to)
		}
	}

	for _, dir := range dirs {
		pkg := packagesByDir[dir]

		// Edge (a): a directed import-dependency edge for each intra-module import. On
		// its own this collapses only import cycles; acyclic imports stay separate
		// targets joined by `deps`.
		for _, imported := range importedPackages(pkg).SortedSlice() {
			for _, owner := range packageToDirs[imported] {
				addEdge(dir, owner)
			}
		}

		// Edge (b): referencing another directory's `internal` symbol forces both into
		// one Kotlin module. "Same module" is symmetric, so the edge is added in BOTH
		// directions -- that 2-cycle is what collapses an otherwise-acyclic internal
		// reference into a single group. (The forward direction usually coincides with
		// an edge (a) import; the back-edge is what edge (b) contributes.)
		if pkg.ImportedClasses != nil {
			for _, cn := range pkg.ImportedClasses.SortedSlice() {
				for _, owner := range internalOwners[cn.FullyQualifiedClassName()] {
					addEdge(dir, owner)
					addEdge(owner, dir)
				}
			}
		}
	}

	groups := buildGroups(tarjanSCC(dirs, adjacency), packagesByDir)
	assignNames(groups)

	g := &Graph{
		groups:         groups,
		dirToGroup:     make(map[string]*Group),
		packageToGroup: make(map[types.PackageName]*Group),
	}
	for _, grp := range groups {
		for _, dir := range grp.Dirs {
			g.dirToGroup[dir] = grp
		}
		for _, pkg := range grp.Packages.SortedSlice() {
			g.packageToGroup[pkg] = grp
		}
	}

	// The condensation of an SCC decomposition is always a DAG; assert it as a guard
	// against a misclassified edge rather than trusting it silently.
	if err := g.assertAcyclic(adjacency); err != nil {
		return nil, err
	}
	return g, nil
}

// importedPackages is the set of packages a package depends on: explicit package
// imports plus the package of every imported class.
func importedPackages(pkg *java.Package) *sorted_set.SortedSet[types.PackageName] {
	result := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
	if pkg.ImportedPackagesWithoutSpecificClasses != nil {
		for _, p := range pkg.ImportedPackagesWithoutSpecificClasses.SortedSlice() {
			result.Add(p)
		}
	}
	if pkg.ImportedClasses != nil {
		for _, cn := range pkg.ImportedClasses.SortedSlice() {
			result.Add(cn.PackageName())
		}
	}
	return result
}

func buildGroups(components [][]string, packagesByDir map[string]*java.Package) []*Group {
	groups := make([]*Group, 0, len(components))
	for _, comp := range components {
		packages := sorted_set.NewSortedSetFn([]types.PackageName{}, types.PackageNameLess)
		for _, dir := range comp {
			packages.Add(packagesByDir[dir].Name)
		}
		groups = append(groups, &Group{Dirs: comp, Packages: packages})
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Dirs[0] < groups[j].Dirs[0] })
	return groups
}

// tarjanSCC returns the strongly-connected components as lexicographically-sorted
// directory slices. Nodes and successors are visited in sorted order so the result
// is deterministic.
func tarjanSCC(dirs []string, adjacency map[string]*sorted_set.SortedSet[string]) [][]string {
	index := make(map[string]int, len(dirs))
	lowlink := make(map[string]int, len(dirs))
	onStack := make(map[string]bool, len(dirs))
	var stack []string
	counter := 0
	var components [][]string

	var strongConnect func(v string)
	strongConnect = func(v string) {
		index[v] = counter
		lowlink[v] = counter
		counter++
		stack = append(stack, v)
		onStack[v] = true

		for _, w := range adjacency[v].SortedSlice() {
			if _, seen := index[w]; !seen {
				strongConnect(w)
				if lowlink[w] < lowlink[v] {
					lowlink[v] = lowlink[w]
				}
			} else if onStack[w] && index[w] < lowlink[v] {
				lowlink[v] = index[w]
			}
		}

		if lowlink[v] == index[v] {
			var component []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				component = append(component, w)
				if w == v {
					break
				}
			}
			sort.Strings(component)
			components = append(components, component)
		}
	}

	for _, v := range dirs {
		if _, seen := index[v]; !seen {
			strongConnect(v)
		}
	}
	return components
}

// assignNames gives each group a deterministic, module-unique name. The default is
// the basename of its smallest member directory; collisions (common in module mode,
// e.g. ".../actions/ui" and ".../eventing/ui") are broken by prepending parent path
// segments until every name is unique.
func assignNames(groups []*Group) {
	segments := make([][]string, len(groups))
	depth := make([]int, len(groups))
	for i, g := range groups {
		segments[i] = strings.Split(g.Dirs[0], "/")
		depth[i] = 1
	}
	nameOf := func(i int) string {
		s := segments[i]
		d := depth[i]
		if d > len(s) {
			d = len(s)
		}
		return strings.Join(s[len(s)-d:], "_")
	}
	for {
		counts := make(map[string]int)
		for i := range groups {
			counts[nameOf(i)]++
		}
		changed := false
		for i := range groups {
			if counts[nameOf(i)] > 1 && depth[i] < len(segments[i]) {
				depth[i]++
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	for i, g := range groups {
		g.Name = nameOf(i)
	}
}

// assertAcyclic verifies the group condensation has no cycle, reporting the offending
// path if it does. Mirrors the white/gray/black DFS used elsewhere in the repo.
func (g *Graph) assertAcyclic(adjacency map[string]*sorted_set.SortedSet[string]) error {
	condensation := make(map[string]*sorted_set.SortedSet[string], len(g.groups))
	names := make([]string, 0, len(g.groups))
	for _, grp := range g.groups {
		condensation[grp.Name] = sorted_set.NewSortedSet([]string{})
		names = append(names, grp.Name)
	}
	sort.Strings(names)
	for _, grp := range g.groups {
		for _, dir := range grp.Dirs {
			for _, dep := range adjacency[dir].SortedSlice() {
				if target := g.dirToGroup[dep]; target != nil && target != grp {
					condensation[grp.Name].Add(target.Name)
				}
			}
		}
	}

	const (
		white = iota
		gray
		black
	)
	color := make(map[string]int, len(names))
	var path []string
	var visit func(n string) error
	visit = func(n string) error {
		color[n] = gray
		path = append(path, n)
		for _, m := range condensation[n].SortedSlice() {
			switch color[m] {
			case gray:
				return fmt.Errorf(
					"scc: group dependency graph has a cycle (a misclassified edge): %s",
					strings.Join(append(append([]string{}, path...), m), " -> "))
			case white:
				if err := visit(m); err != nil {
					return err
				}
			}
		}
		path = path[:len(path)-1]
		color[n] = black
		return nil
	}
	for _, n := range names {
		if color[n] == white {
			if err := visit(n); err != nil {
				return err
			}
		}
	}
	return nil
}

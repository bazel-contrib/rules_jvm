package gazelle

import (
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

// populateProductionAssociatesAttr preserves Kotlin's module-wide `internal` across the
// fine-grained per-package targets that scc granularity produces. Under that granularity a
// module's libraries all share one Bazel package, so:
//
//   - a non-leaf (one with same-package kt_jvm_library deps) friends those deps via
//     `associates`, moving them out of `deps`, and adopts their shared module_name -- so it
//     must NOT also set module_name (rules_kotlin forbids both); and
//   - a leaf (no same-package kt deps) pins an explicit module_name (the package path, a
//     stable id every target in the module shares) so the non-leaves that associate it agree
//     on one Kotlin module.
//
// rules_kotlin requires all of a target's associates to share one module_name and analyses
// deps bottom-up, so the leaves' shared module_name propagates to every target. Only scc
// granularity (many packages, one module) is touched; package granularity is unchanged.
func (jr *Resolver) populateProductionAssociatesAttr(c *config.Config, r *rule.Rule, from label.Label) {
	pc := c.Exts[languageName].(javaconfig.Configs)[from.Pkg]
	if pc == nil || pc.ModuleGranularity() != "scc" {
		return
	}

	var associates, deps []string
	for _, dep := range r.AttrStrings("deps") {
		parsed, err := label.Parse(dep)
		if err != nil {
			deps = append(deps, dep)
			continue
		}
		abs := parsed.Abs(from.Repo, from.Pkg)
		if abs.Repo == from.Repo && abs.Pkg == from.Pkg && jr.lang.kotlinLibraries[label.New("", abs.Pkg, abs.Name).String()] {
			associates = append(associates, dep)
		} else {
			deps = append(deps, dep)
		}
	}

	if r.ShouldKeep() {
		return
	}
	setGeneratedStringListAttr(r, "associates", associates)

	associateSet := newLabelIdentitySet(r.AttrStrings("associates"), from)
	if len(associateSet) > 0 {
		// Non-leaf: friend same-module deps; it adopts their module_name, so it must not set one.
		r.DelAttr("module_name")
		filteredDeps := make([]string, 0, len(deps))
		for _, dep := range deps {
			if _, found := associateSet[labelIdentityOf(dep, from)]; !found {
				filteredDeps = append(filteredDeps, dep)
			}
		}
		if len(filteredDeps) == 0 {
			r.DelAttr("deps")
		} else {
			r.SetAttr("deps", filteredDeps)
		}
		return
	}

	// Leaf: pin the shared module name (the package path).
	r.SetAttr("module_name", strings.ReplaceAll(from.Pkg, "/", "_"))
}

type labelIdentity struct {
	label   label.Label
	invalid string
}

func labelIdentityOf(value string, from label.Label) labelIdentity {
	parsed, err := label.Parse(value)
	if err != nil {
		return labelIdentity{invalid: value}
	}
	return labelIdentity{label: parsed.Abs(from.Repo, from.Pkg)}
}

func newLabelIdentitySet(values []string, from label.Label) map[labelIdentity]struct{} {
	set := make(map[labelIdentity]struct{}, len(values))
	for _, value := range values {
		set[labelIdentityOf(value, from)] = struct{}{}
	}
	return set
}

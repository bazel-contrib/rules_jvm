package gazelle

import (
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
	bzl "github.com/bazelbuild/buildtools/build"
)

// setGeneratedStringListAttr replaces a generated list while retaining entries
// explicitly marked with Gazelle's # keep comment.
func setGeneratedStringListAttr(r *rule.Rule, attrName string, generated []string) bool {
	if attrShouldKeep(r, attrName) {
		return false
	}
	return setStringListAttr(r, attrName, r.Attr(attrName), generated)
}

func setStringListAttr(r *rule.Rule, attrName string, existing bzl.Expr, generated []string) bool {
	if r.ShouldKeep() || (existing != nil && rule.ShouldKeep(existing)) {
		return false
	}

	merged := mergeKeptStringListAttr(existing, generated)
	// DelAttr is required before SetAttr to replace an attribute loaded from BUILD.
	r.DelAttr(attrName)
	if len(merged.List) > 0 {
		r.SetAttr(attrName, merged)
	}
	return true
}

func mergeKeptStringListAttr(existing bzl.Expr, generated []string) *bzl.ListExpr {
	keptByValue := make(map[string]*bzl.StringExpr)
	keptInOrder := make([]*bzl.StringExpr, 0)
	if list, ok := existing.(*bzl.ListExpr); ok {
		for _, expr := range list.List {
			value, ok := expr.(*bzl.StringExpr)
			if !ok || !rule.ShouldKeep(value) {
				continue
			}
			if _, found := keptByValue[value.Value]; found {
				continue
			}
			keptByValue[value.Value] = value
			keptInOrder = append(keptInOrder, value)
		}
	}

	merged := make([]bzl.Expr, 0, len(generated)+len(keptInOrder))
	seen := make(map[string]bool, len(generated)+len(keptInOrder))
	for _, value := range generated {
		if seen[value] {
			continue
		}
		seen[value] = true
		if kept, ok := keptByValue[value]; ok {
			merged = append(merged, kept)
		} else {
			merged = append(merged, &bzl.StringExpr{Value: value})
		}
	}
	for _, kept := range keptInOrder {
		if !seen[kept.Value] {
			seen[kept.Value] = true
			merged = append(merged, kept)
		}
	}
	return &bzl.ListExpr{List: merged}
}

func copyExistingKeptStringListAttrs(c *config.Config, file *rule.File, generated []*rule.Rule, attrNames ...string) {
	if file == nil {
		return
	}

	existingByKey := make(map[[2]string]*rule.Rule, len(file.Rules))
	for _, existing := range file.Rules {
		existingByKey[stringListRuleKey(c, existing)] = existing
	}
	for _, current := range generated {
		if current.ShouldKeep() {
			continue
		}
		existing := existingByKey[stringListRuleKey(c, current)]
		if existing == nil || existing.ShouldKeep() {
			continue
		}
		for _, attrName := range attrNames {
			oldAttr := existing.Attr(attrName)
			if oldAttr == nil {
				continue
			}
			if attrShouldKeep(existing, attrName) {
				current.SetAttr(attrName, oldAttr)
				*current.AttrComments(attrName) = *existing.AttrComments(attrName)
				continue
			}
			setStringListAttr(current, attrName, oldAttr, current.AttrStrings(attrName))
		}
	}
}

func stringListRuleKey(c *config.Config, r *rule.Rule) [2]string {
	kind := r.Kind()
	if c != nil {
		if mapped := mappedKind(c, kind); mapped != "" {
			kind = mapped
		}
	}
	return [2]string{kind, r.Name()}
}

func attrShouldKeep(r *rule.Rule, attrName string) bool {
	attr := r.Attr(attrName)
	if attr == nil {
		return false
	}
	if rule.ShouldKeep(attr) {
		return true
	}
	comments := r.AttrComments(attrName)
	if comments == nil {
		return false
	}
	probe := &bzl.StringExpr{}
	*probe.Comment() = *comments
	return rule.ShouldKeep(probe)
}

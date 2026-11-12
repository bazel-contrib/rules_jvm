package gazelle

import (
	"reflect"
	"strings"
	"testing"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
	bzl "github.com/bazelbuild/buildtools/build"
)

func TestSetGeneratedStringListAttrPreservesKeptEntries(t *testing.T) {
	file, err := rule.LoadData("BUILD.bazel", "", []byte(`java_library(
	name = "example",
	visibility = [
		"//:stale",
		"//:manual",  # keep
		"//:generated",  # keep
	],
)`))
	if err != nil {
		t.Fatal(err)
	}
	r := file.Rules[0]
	var kept *bzl.StringExpr
	for _, expr := range r.Attr("visibility").(*bzl.ListExpr).List {
		if s, ok := expr.(*bzl.StringExpr); ok && s.Value == "//:manual" {
			kept = s
		}
	}

	setGeneratedStringListAttr(r, "visibility", []string{"//:new", "//:generated"})

	if got, want := r.AttrStrings("visibility"), []string{"//:new", "//:generated", "//:manual"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("visibility = %v, want %v", got, want)
	}
	for _, expr := range r.Attr("visibility").(*bzl.ListExpr).List {
		if s, ok := expr.(*bzl.StringExpr); ok && s.Value == "//:manual" {
			if s != kept {
				t.Error("kept visibility entry was replaced instead of reused")
			}
			if !rule.ShouldKeep(s) {
				t.Error("kept visibility comment was lost")
			}
		}
	}
}

func TestSetGeneratedStringListAttrIsStableAcrossRegeneration(t *testing.T) {
	file, err := rule.LoadData("BUILD.bazel", "", []byte(`java_library(
	name = "example",
	associates = [
		":manual",  # keep
		":stale",
	],
)`))
	if err != nil {
		t.Fatal(err)
	}

	setGeneratedStringListAttr(file.Rules[0], "associates", []string{":current"})
	file.Sync()
	first := string(bzl.Format(file.File))

	regenerated, err := rule.LoadData("BUILD.bazel", "", []byte(first))
	if err != nil {
		t.Fatal(err)
	}
	setGeneratedStringListAttr(regenerated.Rules[0], "associates", []string{":current"})
	regenerated.Sync()
	second := string(bzl.Format(regenerated.File))

	if first != second {
		t.Fatalf("regeneration changed BUILD output:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if got, want := regenerated.Rules[0].AttrStrings("associates"), []string{":current", ":manual"}; !reflect.DeepEqual(got, want) {
		t.Errorf("associates = %v, want %v", got, want)
	}
	if strings.Contains(second, ":stale") {
		t.Error("stale unkept associate survived regeneration")
	}
}

func TestCopyExistingKeptStringListAttrs(t *testing.T) {
	file, err := rule.LoadData("BUILD.bazel", "", []byte(`java_library(
	name = "example",
	visibility = [
		"//:manual",  # keep
		"//:stale",
	],
)`))
	if err != nil {
		t.Fatal(err)
	}
	generated := rule.NewRule("java_library", "example")
	generated.SetAttr("visibility", []string{"//:generated"})

	copyExistingKeptStringListAttrs(&config.Config{}, file, []*rule.Rule{generated}, "visibility")

	if got, want := generated.AttrStrings("visibility"), []string{"//:generated", "//:manual"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("visibility = %v, want %v", got, want)
	}
	for _, expr := range generated.Attr("visibility").(*bzl.ListExpr).List {
		if s, ok := expr.(*bzl.StringExpr); ok && s.Value == "//:manual" && !rule.ShouldKeep(s) {
			t.Error("copied visibility comment was lost")
		}
	}
}

func TestCopyExistingKeptStringListAttrsMatchesMappedKind(t *testing.T) {
	file, err := rule.LoadData("BUILD.bazel", "", []byte(`custom_java_library(
	name = "example",
	visibility = [
		"//:manual",  # keep
	],
)`))
	if err != nil {
		t.Fatal(err)
	}
	generated := rule.NewRule("java_library", "example")
	generated.SetAttr("visibility", []string{"//:generated"})
	c := &config.Config{KindMap: map[string]config.MappedKind{
		"java_library": {
			FromKind: "java_library",
			KindName: "custom_java_library",
		},
	}}

	copyExistingKeptStringListAttrs(c, file, []*rule.Rule{generated}, "visibility")

	if got, want := generated.AttrStrings("visibility"), []string{"//:generated", "//:manual"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("visibility = %v, want %v", got, want)
	}
	foundKept := false
	for _, expr := range generated.Attr("visibility").(*bzl.ListExpr).List {
		if value, ok := expr.(*bzl.StringExpr); ok && value.Value == "//:manual" {
			foundKept = rule.ShouldKeep(value)
		}
	}
	if !foundKept {
		t.Error("mapped visibility entry lost its keep comment")
	}
}

func TestCopyExistingKeptStringListAttrsCopiesKeptAttribute(t *testing.T) {
	file, err := rule.LoadData("BUILD.bazel", "", []byte(`java_library(
	name = "example",
	visibility = ["//:manual"],  # keep
)`))
	if err != nil {
		t.Fatal(err)
	}
	generated := rule.NewRule("java_library", "example")
	generated.SetAttr("visibility", []string{"//:generated"})

	copyExistingKeptStringListAttrs(&config.Config{}, file, []*rule.Rule{generated}, "visibility")

	if got, want := generated.AttrStrings("visibility"), []string{"//:manual"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("visibility = %v, want %v", got, want)
	}
	if comments := generated.AttrComments("visibility"); comments == nil || len(comments.Suffix) == 0 {
		t.Error("copied visibility attribute lost its keep comment")
	}
}

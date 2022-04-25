package main

import (
	"testing"

	pb "github.com/bazel-contrib/rules_jvm/java/gazelle/private/javaparser/proto/gazelle/java/javaparser/v0"
	"github.com/google/go-cmp/cmp"
)

func Test_classNamesFromFiles(t *testing.T) {
	want := []string{"Foo", "Bar"}
	got := classNamesFromFiles([]string{"Foo.java", "Bar.java"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("classNamesFromFiles() mismatch (-want +got):\n%s", diff)
	}
}

func Test_filterImports(t *testing.T) {
	pkg := &pb.Package{
		Name: "com.example.a.b",
		Imports: []string{
			"com.example.a.b.Foo",        // same pkg, included class name: delete
			"com.example.a.b.Bar",        // same pkg, included class name: delete
			"com.example.a.b.Bar.SubBar", // same pkg, nested class, included class name: delete
			"com.example.a.b.Baz",        // same pkg, not included class name: keep
			"com.example.a.b.Baz.SubBaz", // same pkg, nested class, not included class name: keep
			"com.example.a.b.c.Foo",      // different pkg: keep
			"com.example.a.Foo",          // different pkg: keep
			"com.another.a.b.Foo",        // different pkg: keep

		},
	}
	classNames := []string{
		"Foo",
		"Bar",
	}
	want := []string{
		"com.example.a.b.Baz",
		"com.example.a.b.Baz.SubBaz",
		"com.example.a.b.c.Foo",
		"com.example.a.Foo",
		"com.another.a.b.Foo",
	}

	got := filterImports(pkg, classNames)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("filterImports() mismatch (-want +got):\n%s", diff)
	}
}

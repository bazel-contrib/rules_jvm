package javaconfig_test

import (
	"testing"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/javaconfig"
)

func TestDefaultTestSuffixes(t *testing.T) {
	for file, want := range map[string]bool{
		"":                    false,
		"Foo.java":            false,
		"Foo.class":           false,
		"TestFoo.java":        false,
		"FooTest.java":        true,
		"BarTest.java":        true,
		"FooTestCase.java":    false,
		"FooIT.java":          false,
		"FooTestContext.java": false,
		"Test.java":           true,
		"FooTest.cpp":         false,
	} {
		t.Run(file, func(t *testing.T) {
			config := javaconfig.New("/tmp")
			if got := config.IsJavaTestFile(file); got != want {
				t.Fatalf("%s: want %v got %v", file, want, got)
			}
		})
	}
}

func TestOverriddenTestSuffixes(t *testing.T) {
	for file, want := range map[string]bool{
		"":                    false,
		"Foo.java":            false,
		"Foo.class":           false,
		"TestFoo.java":        false,
		"FooTest.java":        true,
		"BarTest.java":        false,
		"FooTestCase.java":    false,
		"AnotherFooIT.java":   true,
		"FooIT.java":          true,
		"FooTestContext.java": false,
		"Test.java":           false,
		"FooTest.cpp":         false,
	} {
		t.Run(file, func(t *testing.T) {
			config := javaconfig.New("/tmp")
			config.SetJavaTestFileSuffixes("FooTest.java,IT.java")
			if got := config.IsJavaTestFile(file); got != want {
				t.Fatalf("%s: want %v got %v", file, want, got)
			}
		})
	}
}

// A child package must inherit its parent's java_generate_proto setting, like
// every other directive. Guards against NewChild() resetting generateProto
// instead of copying it from the parent.
func TestGenerateProtoInheritsToChild(t *testing.T) {
	parent := javaconfig.New("/tmp")
	parent.SetGenerateProto(false)

	if child := parent.NewChild(); child.GenerateProto() {
		t.Fatalf("child did not inherit generateProto=false from parent; got true")
	}
}

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

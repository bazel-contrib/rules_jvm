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

func TestMapLibraryName(t *testing.T) {
	tests := []struct {
		name       string
		convention string
		dirname    string
		want       string
	}{
		{"default", "{dirname}", "hello", "hello"},
		{"prefix", "lib_{dirname}", "hello", "lib_hello"},
		{"suffix", "{dirname}_lib", "hello", "hello_lib"},
		{"no placeholder", "mylib", "hello", "mylib"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := javaconfig.New("/tmp")
			if tt.convention != "{dirname}" {
				config.SetLibraryNamingConvention(tt.convention)
			}
			if got := config.MapLibraryName(tt.dirname); got != tt.want {
				t.Fatalf("MapLibraryName(%q) = %q, want %q", tt.dirname, got, tt.want)
			}
		})
	}
}

func TestMapTestSuiteName(t *testing.T) {
	tests := []struct {
		name       string
		convention string
		dirname    string
		isModule   bool
		want       string
	}{
		{"default package mode", "", "hello", false, "hello"},
		{"default module mode", "", "hello", true, "hello-tests"},
		{"custom package mode", "{dirname}_tests", "hello", false, "hello_tests"},
		{"custom module mode", "{dirname}_tests", "hello", true, "hello_tests"},
		{"no placeholder", "all_tests", "hello", false, "all_tests"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := javaconfig.New("/tmp")
			if tt.convention != "" {
				config.SetTestSuiteNamingConvention(tt.convention)
			}
			if got := config.MapTestSuiteName(tt.dirname, tt.isModule); got != tt.want {
				t.Fatalf("MapTestSuiteName(%q, %v) = %q, want %q", tt.dirname, tt.isModule, got, tt.want)
			}
		})
	}
}

func TestNamingConventionInheritance(t *testing.T) {
	parent := javaconfig.New("/tmp")
	parent.SetLibraryNamingConvention("lib_{dirname}")
	parent.SetTestSuiteNamingConvention("{dirname}_tests")

	child := parent.NewChild()

	if got := child.MapLibraryName("hello"); got != "lib_hello" {
		t.Fatalf("child MapLibraryName = %q, want %q", got, "lib_hello")
	}
	if got := child.MapTestSuiteName("hello", false); got != "hello_tests" {
		t.Fatalf("child MapTestSuiteName = %q, want %q", got, "hello_tests")
	}
}

func TestSetNamingConventionValidation(t *testing.T) {
	config := javaconfig.New("/tmp")

	if err := config.SetLibraryNamingConvention(""); err == nil {
		t.Fatal("expected error for empty library naming convention")
	}
	if err := config.SetTestSuiteNamingConvention(""); err == nil {
		t.Fatal("expected error for empty test suite naming convention")
	}
	if err := config.SetLibraryNamingConvention("lib_{dirname}"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := config.SetTestSuiteNamingConvention("{dirname}_tests"); err != nil {
		t.Fatalf("unexpected error: %v", err)
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

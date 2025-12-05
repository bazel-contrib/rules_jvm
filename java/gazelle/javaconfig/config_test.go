package javaconfig_test

import (
	"os"
	"path/filepath"
	"sort"
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

func TestPathsForPackage(t *testing.T) {
	tests := []struct {
		name        string
		searchPaths []string
		pkg         string
		want        []string
	}{
		{
			name:        "no search paths configured",
			searchPaths: nil,
			pkg:         "com.example.util",
			want:        nil,
		},
		{
			name:        "simple directory without package prefix",
			searchPaths: []string{"src/main/java"},
			pkg:         "com.example.util",
			want:        []string{"src/main/java/com/example/util"},
		},
		{
			name:        "directory with package prefix - exact match",
			searchPaths: []string{"third_party/example package=com.example"},
			pkg:         "com.example",
			want:        []string{"third_party/example"},
		},
		{
			name:        "directory with package prefix - subpackage",
			searchPaths: []string{"third_party/example package=com.example"},
			pkg:         "com.example.util",
			want:        []string{"third_party/example/util"},
		},
		{
			name:        "directory with package prefix - no match",
			searchPaths: []string{"third_party/example package=com.example"},
			pkg:         "org.other.lib",
			want:        nil,
		},
		{
			name:        "multiple search paths",
			searchPaths: []string{"src/main/java", "src/generated/java"},
			pkg:         "com.example.util",
			want:        []string{"src/main/java/com/example/util", "src/generated/java/com/example/util"},
		},
		{
			name:        "mixed search paths with and without package prefix",
			searchPaths: []string{"src/main/java", "third_party/example package=com.example"},
			pkg:         "com.example.util",
			want:        []string{"src/main/java/com/example/util", "third_party/example/util"},
		},
		{
			name:        "deeply nested package",
			searchPaths: []string{"src/main/java"},
			pkg:         "com.example.internal.util.helpers",
			want:        []string{"src/main/java/com/example/internal/util/helpers"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := javaconfig.New("/tmp")
			for _, sp := range tt.searchPaths {
				if err := config.AddSearchPath(sp); err != nil {
					t.Fatalf("AddSearchPath(%q) failed: %v", sp, err)
				}
			}
			got := config.PathsForPackage(tt.pkg)
			if len(got) != len(tt.want) {
				t.Fatalf("PathsForPackage(%q) = %v, want %v", tt.pkg, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("PathsForPackage(%q)[%d] = %q, want %q", tt.pkg, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDiscoverMavenLayout(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "maven-layout-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directory structure
	dirs := []string{
		"src/main/java",
		"src/main/kotlin",
		"src/test/java",
		"src/integrationTest/java",
		"moduleA/src/main/java",
		"moduleA/src/test/kotlin",
		"parent/child/src/main/java",
		"not-a-sourceset/java", // missing src parent
		"src/main/resources",   // not java or kotlin
		"deep/nested/module/src/main/java",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		directive string
		want      []string
	}{
		{
			name:      "depth 0 finds root sourcesets only",
			directive: "depth=0",
			want: []string{
				"src/integrationTest/java",
				"src/main/java",
				"src/main/kotlin",
				"src/test/java",
			},
		},
		{
			name:      "depth 1 finds moduleA too",
			directive: "depth=1",
			want: []string{
				"moduleA/src/main/java",
				"moduleA/src/test/kotlin",
				"src/integrationTest/java",
				"src/main/java",
				"src/main/kotlin",
				"src/test/java",
			},
		},
		{
			name:      "depth 2 finds parent/child module",
			directive: "depth=2",
			want: []string{
				"moduleA/src/main/java",
				"moduleA/src/test/kotlin",
				"parent/child/src/main/java",
				"src/integrationTest/java",
				"src/main/java",
				"src/main/kotlin",
				"src/test/java",
			},
		},
		{
			name:      "depth 4 finds deeply nested",
			directive: "depth=4",
			want: []string{
				"deep/nested/module/src/main/java",
				"moduleA/src/main/java",
				"moduleA/src/test/kotlin",
				"parent/child/src/main/java",
				"src/integrationTest/java",
				"src/main/java",
				"src/main/kotlin",
				"src/test/java",
			},
		},
		{
			name:      "empty directive uses default depth",
			directive: "",
			want: []string{
				"deep/nested/module/src/main/java",
				"moduleA/src/main/java",
				"moduleA/src/test/kotlin",
				"parent/child/src/main/java",
				"src/integrationTest/java",
				"src/main/java",
				"src/main/kotlin",
				"src/test/java",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := javaconfig.New(tmpDir)
			got, err := config.DiscoverMavenLayout(tt.directive)
			if err != nil {
				t.Fatalf("DiscoverMavenLayout(%q) failed: %v", tt.directive, err)
			}
			sort.Strings(got)

			if len(got) != len(tt.want) {
				t.Fatalf("DiscoverMavenLayout(%q) = %v, want %v", tt.directive, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("DiscoverMavenLayout(%q)[%d] = %q, want %q", tt.directive, i, got[i], tt.want[i])
				}
			}

			// Verify search paths were added
			searchPaths := config.SearchPaths()
			if len(searchPaths) != len(tt.want) {
				t.Fatalf("Expected %d search paths, got %d", len(tt.want), len(searchPaths))
			}
		})
	}
}

func TestIsSearchExcluded(t *testing.T) {
	tests := []struct {
		name     string
		excludes []string
		dir      string
		want     bool
	}{
		{
			name:     "no excludes",
			excludes: nil,
			dir:      "src/main/java",
			want:     false,
		},
		{
			name:     "exact match",
			excludes: []string{"build/generated"},
			dir:      "build/generated",
			want:     true,
		},
		{
			name:     "subdirectory excluded",
			excludes: []string{"build/generated"},
			dir:      "build/generated/src/main/java",
			want:     true,
		},
		{
			name:     "no match - different path",
			excludes: []string{"build/generated"},
			dir:      "src/main/java",
			want:     false,
		},
		{
			name:     "multiple excludes - second matches",
			excludes: []string{"build/generated", "third_party"},
			dir:      "third_party/vendored/src/main/java",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := javaconfig.New("/tmp")
			for _, exclude := range tt.excludes {
				config.AddSearchExclude(exclude)
			}
			got := config.IsSearchExcluded(tt.dir)
			if got != tt.want {
				t.Fatalf("IsSearchExcluded(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestDiscoverMavenLayoutWithExcludes(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "maven-layout-exclude-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test directory structure
	dirs := []string{
		"src/main/java",
		"src/test/java",
		"build/generated/src/main/java",
		"moduleA/src/main/java",
		"third_party/vendored/src/main/java",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name     string
		excludes []string
		want     []string
	}{
		{
			name:     "no excludes",
			excludes: nil,
			want: []string{
				"build/generated/src/main/java",
				"moduleA/src/main/java",
				"src/main/java",
				"src/test/java",
				"third_party/vendored/src/main/java",
			},
		},
		{
			name:     "exclude build/generated",
			excludes: []string{"build/generated"},
			want: []string{
				"moduleA/src/main/java",
				"src/main/java",
				"src/test/java",
				"third_party/vendored/src/main/java",
			},
		},
		{
			name:     "exclude multiple directories",
			excludes: []string{"build/generated", "third_party"},
			want: []string{
				"moduleA/src/main/java",
				"src/main/java",
				"src/test/java",
			},
		},
		{
			name:     "exclude with parent directory",
			excludes: []string{"build"},
			want: []string{
				"moduleA/src/main/java",
				"src/main/java",
				"src/test/java",
				"third_party/vendored/src/main/java",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := javaconfig.New(tmpDir)
			for _, exclude := range tt.excludes {
				config.AddSearchExclude(exclude)
			}
			got, err := config.DiscoverMavenLayout("")
			if err != nil {
				t.Fatalf("DiscoverMavenLayout failed: %v", err)
			}
			sort.Strings(got)

			if len(got) != len(tt.want) {
				t.Fatalf("DiscoverMavenLayout() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("DiscoverMavenLayout()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestDiscoverMavenLayoutOnlyRunsOnce(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "maven-layout-once-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple structure
	if err := os.MkdirAll(filepath.Join(tmpDir, "src/main/java"), 0755); err != nil {
		t.Fatal(err)
	}

	config := javaconfig.New(tmpDir)

	// First call should discover
	got1, err := config.DiscoverMavenLayout("")
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if len(got1) != 1 {
		t.Fatalf("First call: expected 1 path, got %d", len(got1))
	}

	// Second call should return nil (already discovered)
	got2, err := config.DiscoverMavenLayout("")
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if got2 != nil {
		t.Fatalf("Second call: expected nil, got %v", got2)
	}

	// Search paths should still be correct
	if len(config.SearchPaths()) != 1 {
		t.Fatalf("Expected 1 search path, got %d", len(config.SearchPaths()))
	}
}

func TestAddSearchPathParsing(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:  "simple directory",
			value: "src/main/java",
		},
		{
			name:  "directory with package",
			value: "third_party/example package=com.example",
		},
		{
			name:    "unknown option",
			value:   "src/main/java foo=bar",
			wantErr: true,
		},
		{
			name:    "empty value",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := javaconfig.New("/tmp")
			err := config.AddSearchPath(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(config.SearchPaths()) != 1 {
				t.Fatalf("expected 1 search path, got %d", len(config.SearchPaths()))
			}
		})
	}
}

func TestDiscoverMavenLayoutParsing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "maven-layout-parse-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:  "empty value uses default",
			value: "",
		},
		{
			name:  "explicit depth",
			value: "depth=3",
		},
		{
			name:  "depth zero",
			value: "depth=0",
		},
		{
			name:    "negative depth",
			value:   "depth=-1",
			wantErr: true,
		},
		{
			name:    "invalid depth",
			value:   "depth=abc",
			wantErr: true,
		},
		{
			name:    "unknown option",
			value:   "foo=bar",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := javaconfig.New(tmpDir)
			_, err := config.DiscoverMavenLayout(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

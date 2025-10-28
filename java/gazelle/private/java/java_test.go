package java

import (
	"testing"
)

func TestIsTestPackage(t *testing.T) {
	tests := map[string]bool{
		"":                                             false,
		"java/client/src/org/openqa/selenium":          false,
		"java/client/test/org/openqa/selenium":         true,
		"java/com/example/platform/githubhook":         false,
		"javatests/com/example/platform/githubhook":    true,
		"project1/src/main/java/com/example/myproject": false,
		"project1/src/test/java/com/example/myproject": true,
		"project1/src/integrationTest/more":            true,
		"src/main/java/com/example/myproject":          false,
		"src/test/java/com/example/myproject":          true,
		"src/main/com/example/test":                    false,
		"src/main/com/example/perftest":                false,
		"test-utils/src/main/com/example/project":      false,
		"foo/bar/test":                                 true,
		"project1/testutils/src/main":                  false,
	}

	for pkg, want := range tests {
		t.Run(pkg, func(t *testing.T) {
			if got := IsTestPackage(pkg); got != want {
				t.Errorf("isTest() = %v, want %v", got, want)
			}
		})
	}
}

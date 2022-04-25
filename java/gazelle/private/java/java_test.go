package java

import (
	"path/filepath"
	"testing"
)

func TestIsTestPath(t *testing.T) {
	tests := map[string]bool{
		"": false,
		"java/client/src/org/openqa/selenium/WebDriver.java":                  false,
		"java/client/test/org/openqa/selenium/WebElementTest.java":            true,
		"java/com/example/platform/githubhook/GithubHookModule.java":          false,
		"javatests/com/example/platform/githubhook/GithubHookModuleTest.java": true,
		"project1/src/main/java/com/example/myproject/App.java":               false,
		"project1/src/test/java/com/example/myproject/TestApp.java":           true,
		"project1/src/integrationTest/more/Things.java":                       true,
		"src/main/java/com/example/myproject/App.java":                        false,
		"src/test/java/com/example/myproject/TestApp.java":                    true,
		"src/main/com/example/perftest/Query.java":                            false,
		"src/main/com/example/perftest/TestBase.java":                         false,
		"test-utils/src/main/com/example/project/App.java":                    false,
	}

	for path, want := range tests {
		t.Run(path, func(t *testing.T) {
			dir := filepath.Dir(path)
			if got := IsTestPath(dir); got != want {
				t.Errorf("isTest() = %v, want %v", got, want)
			}
		})
	}
}

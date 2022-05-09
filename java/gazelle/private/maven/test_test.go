package maven

import "testing"

func TestIsTestFile(t *testing.T) {
	tests := map[string]bool{
		"":                    false,
		"Foo.java":            false,
		"Foo.class":           false,
		"TestFoo.java":        true,
		"FooTest.java":        true,
		"FooTestCase.java":    true,
		"FooIT.java":          false,
		"FooTestContext.java": false,
	}
	for filename, want := range tests {
		t.Run(filename, func(t *testing.T) {
			if got := IsTestFile(filename); got != want {
				t.Errorf("IsTestFile() = %v, want %v", got, want)
			}
		})
	}
}

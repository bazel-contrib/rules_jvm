package gazelle

import (
	"testing"
)

func TestDoneGeneratingRules_NilParser(t *testing.T) {
	lang := NewLanguage().(*javaLang)

	// Simulate the case where no Java files were found (parser was never initialized)
	lang.parser = nil

	// This should not panic
	lang.DoneGeneratingRules()
}

package maven

import "regexp"

var testPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^Test.*\.java$`),
	regexp.MustCompile(`^.*Test\.java$`),
	regexp.MustCompile(`^.*Tests\.java`),
	regexp.MustCompile(`^.*TestCase\.java$`),

	// SystemVariableTestContext.java
}

// IsTestFile returns whether a given filename is test class.
//
// See https://maven.apache.org/surefire/maven-surefire-plugin/examples/inclusion-exclusion.html
func IsTestFile(filename string) bool {
	for _, p := range testPatterns {
		if p.MatchString(filename) {
			return true
		}
	}
	return false
}

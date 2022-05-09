package java

import "strings"

// IsTestPath tries to detect if the directory would contain test files of not.
func IsTestPath(dir string) bool {
	if strings.HasPrefix(dir, "javatests/") {
		return true
	}

	if strings.Contains(dir, "src/") {
		afterSrc := strings.SplitAfterN(dir, "src/", 2)[1]
		firstDir := strings.Split(afterSrc, "/")[0]
		if strings.HasSuffix(strings.ToLower(firstDir), "test") {
			return true
		}
	}

	return strings.Contains(dir, "/test/")
}

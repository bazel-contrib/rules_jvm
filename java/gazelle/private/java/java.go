package java

import (
	"strings"

	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

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

var stdlibPrefixes = []types.PackageName{
	types.NewPackageName("com.sun.management"),
	types.NewPackageName("com.sun.net.httpserver"),
	types.NewPackageName("java"),
	types.NewPackageName("javax.annotation.security"),
	types.NewPackageName("javax.crypto"),
	types.NewPackageName("javax.management"),
	types.NewPackageName("javax.naming"),
	types.NewPackageName("javax.net"),
	types.NewPackageName("javax.security"),
	types.NewPackageName("javax.xml"),
	types.NewPackageName("jdk"),
	types.NewPackageName("org.w3c.dom"),
	types.NewPackageName("org.xml.sax"),
	types.NewPackageName("sun"),
}

func IsStdlib(imp types.PackageName) bool {
	for _, prefix := range stdlibPrefixes {
		if types.PackageNamesHasPrefix(imp, prefix) {
			return true
		}
	}
	return false
}

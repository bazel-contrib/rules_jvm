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

// This list was derived from a script along the lines of:
// for jmod in ${JAVA_HOME}; do unzip -l "${jmod}" 2>/dev/null; done | grep classes/ | awk '{print $4}' | sed -e 's#^classes/##' -e 's#\.class$##' | xargs -n1 dirname | sort | uniq | sed -e 's#/#.#g'
var stdlibPrefixes = []types.PackageName{
	types.NewPackageName("com.sun"),
	types.NewPackageName("java"),
	types.NewPackageName("javax.accessibility"),
	types.NewPackageName("javax.annotation.processing"),
	types.NewPackageName("javax.annotation.security"),
	types.NewPackageName("javax.crypto"),
	types.NewPackageName("javax.imageio"),
	types.NewPackageName("javax.lang.model"),
	types.NewPackageName("javax.management"),
	types.NewPackageName("javax.naming"),
	types.NewPackageName("javax.net"),
	types.NewPackageName("javax.print"),
	types.NewPackageName("javax.rmi.ssl"),
	types.NewPackageName("javax.security"),
	types.NewPackageName("javax.script"),
	types.NewPackageName("javax.smartcardio"),
	types.NewPackageName("javax.sound"),
	types.NewPackageName("javax.sql"),
	types.NewPackageName("javax.swing"),
	types.NewPackageName("javax.tools"),
	types.NewPackageName("javax.transaction.xa"),
	types.NewPackageName("javax.xml"),
	types.NewPackageName("jdk"),
	types.NewPackageName("netscape.javascript"),
	types.NewPackageName("org.ietf.jgss"),
	types.NewPackageName("org.jcp.xml.dsig.internal"),
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

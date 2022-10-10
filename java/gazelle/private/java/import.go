package java

import (
	"strings"
	"unicode"
)

type Import struct {
	Pkg     string
	Classes []string
}

func NewImport(imp string) *Import {
	imp = strings.TrimSuffix(strings.ReplaceAll(imp, "$", ""), ".*")

	parts := strings.Split(imp, ".")
	i := 0
	for ; i < len(parts); i++ {
		if unicode.IsUpper(rune(parts[i][0])) {
			break
		}
		// The "_TESTONLY" comes from the parser to identify cases where there is both a
		// java_library and a java_test_suite in the same file, and a package may need access
		// to the test library. see java/gazelle/generate.go for details.
		if strings.HasPrefix(parts[i], "_") && !strings.HasPrefix(parts[i], "_TESTONLY") {
			break
		}
	}

	return &Import{
		Pkg:     strings.Join(parts[0:i], "."),
		Classes: parts[i:],
	}
}

func (i *Import) Path() string {
	return strings.ReplaceAll(i.Pkg, ".", "/")
}

var stdlibPrefixes = []string{
	"com.sun.management.",
	"com.sun.net.httpserver.",
	"java.",
	"javax.annotation.security.",
	"javax.crypto.",
	"javax.management.",
	"javax.naming.",
	"javax.net.",
	"javax.security.",
	"javax.xml.",
	"jdk.",
	"org.w3c.dom.",
	"org.xml.sax.",
	"sun.",
}

func IsStdlib(imp string) bool {
	for _, prefix := range stdlibPrefixes {
		if strings.HasPrefix(imp, prefix) {
			return true
		}
	}
	return false
}

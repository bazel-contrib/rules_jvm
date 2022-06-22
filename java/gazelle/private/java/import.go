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
	"java.",
	"javax.annotation.security.",
	"javax.crypto.",
	"javax.management.",
	"javax.naming.",
	"javax.net.",
	"javax.security.",
	"javax.xml.",
	"jdk.",
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

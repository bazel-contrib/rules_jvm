package kotlin

import (
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

var kotlinStdLibPrefix = types.NewPackageName("kotlin")

func IsStdlib(imp types.PackageName) bool {
	return types.PackageNamesHasPrefix(imp, kotlinStdLibPrefix)
}

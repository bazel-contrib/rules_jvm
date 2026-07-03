package kotlin

import (
	"github.com/bazel-contrib/rules_jvm/java/gazelle/private/types"
)

var kotlinStdLibPrefix = types.NewPackageName("kotlin")

// kotlinTestPrefix is in the `kotlin` namespace but ships in a separate artifact
// (org.jetbrains.kotlin:kotlin-test), not kotlin-stdlib, so it must resolve to a maven
// dependency rather than being assumed on the classpath.
var kotlinTestPrefix = types.NewPackageName("kotlin.test")

func IsStdlib(imp types.PackageName) bool {
	if types.PackageNamesHasPrefix(imp, kotlinTestPrefix) {
		return false
	}
	return types.PackageNamesHasPrefix(imp, kotlinStdLibPrefix)
}

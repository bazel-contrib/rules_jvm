package workspace.com.gazelle.kotlin.javaparser.generators

// No imports — every referenced type is in the same package but potentially in
// a different Bazel target (split package). The Java parser detects all of these
// via checkFullyQualifiedType's same-package fallback. The Kotlin parser should
// match it.

class SamePackageAllPositions : SomeSuperType() {
    val field: SomeFieldType = throw UnsupportedOperationException()

    fun process(input: SomeParamType): SomeReturnType {
        return throw UnsupportedOperationException()
    }

    fun check(obj: Any): Boolean {
        return obj is SomeMarker
    }

    fun cast(obj: Any): SomeCastTarget {
        return obj as SomeCastTarget
    }
}

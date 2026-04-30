package workspace.com.gazelle.kotlin.javaparser.generators

// No imports — all types are used fully-qualified in type positions.
// KtParser's tryGetFullyQualifiedName only calls getReferencedName() on KtUserType,
// which returns the simple name (last segment). The qualifier chain is never walked,
// so FQN types in these positions are invisible.

class FqnTypePositions {
    // FQN as function parameter type
    fun process(input: com.example.types.InputData): String {
        return input.toString()
    }

    // FQN as function return type
    fun create(): com.example.types.OutputData {
        throw UnsupportedOperationException()
    }

    // FQN as property type
    val config: com.example.config.AppConfig = throw UnsupportedOperationException()

    // FQN in is-check
    fun check(obj: Any): Boolean {
        return obj is com.example.types.Marker
    }

    // FQN in as-cast
    fun cast(obj: Any): Any {
        return obj as com.example.types.Castable
    }
}

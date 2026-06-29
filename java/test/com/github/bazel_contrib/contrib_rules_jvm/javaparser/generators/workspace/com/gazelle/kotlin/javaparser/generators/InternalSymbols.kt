package workspace.com.gazelle.kotlin.javaparser.generators

// Top-level internal declarations: these must be recorded so dependers can be detected as needing
// friend (associate) access.
internal class InternalClass

internal object InternalObject

internal fun internalFunction(): Int = 0

internal const val INTERNAL_CONSTANT = "x"

// Public / private / protected and nested-internal declarations must NOT be recorded: only
// top-level internal symbols matter for package-to-package coupling.
class PublicClass {
  internal fun nestedInternal(): Int = 0
}

fun publicFunction(): Int = 0

private fun privateFunction(): Int = 0

private class PrivateClass

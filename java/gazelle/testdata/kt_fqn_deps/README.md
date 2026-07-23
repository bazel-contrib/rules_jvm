# Kotlin fully qualified name dependencies

Test that Gazelle detects dependencies from fully qualified class references
used directly in Kotlin expressions, without corresponding import statements.

When Kotlin code uses FQN constructor calls like
`com.example.errors.CustomError(e)` instead of importing `CustomError`, the
parser must still recognize the cross-package dependency. This is common when
there are name conflicts (e.g., a custom `InterruptedException` alongside
`java.util.concurrent.InterruptedException`).

The `app/src` package uses a FQN constructor call to `com.example.errors.CustomError`
and should get a `deps` entry for `//errors/src`.

package workspace.com.gazelle.kotlin.javaparser.generators

class CallChainReceivers {
  fun demo() {
    // Call chain: the receiver of the final `.Bar()` selector is `Value.foo(1)`,
    // whose text contains a `.` but is not a fully-qualified identifier (parens).
    // Must not be recorded as a class reference.
    Value.foo(1).Bar()
  }
}

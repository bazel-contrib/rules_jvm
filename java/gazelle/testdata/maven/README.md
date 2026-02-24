Copied from
[bazelbuild/examples/java-maven](https://github.com/bazelbuild/examples/tree/b29794fb55f6714442dd86946c77f8908321a430/java-maven).

This fixture also covers same-package class references without imports in split source roots.
`src/test/java/com/example/myproject/AppTest.java` references `App` as a bare type (`App app = ...`) with no import, which is valid because both files are in `com.example.myproject`.

`src/sample/java/com/example/myproject/SampleOnly.java` is intentionally present to create a second provider of `com.example.myproject` (in addition to `src/main`). That forces ambiguous package-level resolution and requires class-level disambiguation using `imported_classes`.

If the parser drops same-package bare type references, class-level input is missing and the test target can no longer reliably resolve to `//src/main/java/com/example/myproject` (the target that actually provides `App`). With the fix, `com.example.myproject.App` is retained and resolution succeeds.

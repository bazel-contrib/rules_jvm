Tests checking that different configurations of `java_export` targets are handled correctly.

## Subdirectory: `user`

This is the code that makes use of all the dependencies.
If you're reading these tests for the first time, it's the best place to start.

## Subdirectory: `shared_dep`

This test case tests that when two `java_exports` export the same package, an appropriate error message is displayed.
Let's say we have the following repository:

```starlark
java_library(name = "A")
java_export(name = "X", exports = [":B"])
java_export(name = "Y", exports = [":B"])
java_library(name = "B")
```

Then if `:A` wants to depend on `:B`, it has to choose whether it should depend on `:X` or `:Y`.
The implementation picks whatever `java_export` was processed first, based on Gazelle's directory order traversal.
Because this traversal is not deterministic, we can't write deterministic `BUILD.out` files for either `//shared_dep/src/main/java/com/example/shared_dep:BUILD.out` or `//shared_dep/s/m/j/c/e/shared_dep/shared_dep:BUILD.out`.

## Subdirectory: `nested`

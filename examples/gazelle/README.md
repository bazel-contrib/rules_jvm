This example contains java files which don't have BUILD targets for them.

It requires running `bazel run :gazelle` to generate targets which can be built.

There is a test using `genquery` which asserts that the right targets are generated.

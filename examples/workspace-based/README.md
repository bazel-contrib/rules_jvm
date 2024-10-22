# Example Workspace-based project

This is a small example of using the `gazelle` java plugin with a
simple source repository.

## Steps to use

 1. Run `bazel test //...` and notice that there's almost no tests.
 1. `bazel run gazelle` and find the new `BUILD.bazel` files that are created
 1. `bazel test //...`

The final run of `bazel test` will fail, as one of the example tests
is written so it won't pass. Try fixing it!

This example also shows how you can make use of the linters that are
supplied with `contrib_rules_jvm`. There are directives in the
top-level `BUILD.bazel` that tell the gazelle plugin to use the macros
from `contrib_rules_jvm` and not just `rules_java`. The configuration
of the linters can be found in the `WORKSPACE` file.

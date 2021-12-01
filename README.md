# rules_jvm_contrib

Handy rules for working with JVM-based projects in Bazel.

## Freezing Dependencies

At runtime, a handful of dependencies are required by helper classes
in this project. Rather than pollute the default `@maven` workspace,
these are loaded into a `@rules_jvm_contrib_deps` workspace. These
dependencies are loaded using a call to `maven_install`, but we don't
want to force users to remember to load our own dependencies for
us. Instead, to add a new dependency to the project:

1. Update `frozen_deps` in the `WORKSPACE` file
2. Run `./bin/freeze-deps.py`
3. Commit the updated files.
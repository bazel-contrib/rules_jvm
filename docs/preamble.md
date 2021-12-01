# rules_jvm_contrib

Handy rules for working with JVM-based projects in Bazel.

In order to use these in your own projects, in your `WORKSPACE` once
you've used an `http_archive`, you can load all the necessary 
dependencies by:

```starlark
load("@rules_jvm_contrib//:repositories.bzl", "rules_jvm_contrib_deps")

rules_jvm_contrib_deps()

load("@rules_jvm_contrib//:setup.bzl", "rules_jvm_contrib_setup")

rules_jvm_contrib_setup()
```

## Java Rules


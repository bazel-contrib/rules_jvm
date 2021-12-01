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

## Linting

Many of the features in this repo are designed to be exposed via [apple_rules_lint][arl], which provides a framework for integrating linting checks into your builds. To take advantage of this perform the following steps:

```starlark
# In your WORKSPACE, after loading `apple_rules_lint`

load("@apple_rules_lint//lint:setup.bzl", "lint_setup")

lint_setup({
  # Note: this is an example config!
  "java-checkstyle": "@rules_jvm_contrib//java:checkstyle-default-config",
  "java-pmd": "//:pmd-config",
  "java-spotbugs": "@rules_jvm_contrib//java:spotbugs-default-config",
})
```

You are welcome to include all (or none!) of these rules, and linting
is "opt-in": if there's no `lint_setup` call in your repo's
`WORKSPACE` then everything will continue working just fine and no
additional lint tests will be generated.

The linters are configured using specific rules. The mappings are:

| Well known name | Lint config rule |
|-----------------|------------------|
| java-checkstyle | checkstyle_config |
| java-pmd | pmd_ruleset |
| java-spotbugs | spotbugs_config |

## Requirements

These rules require Java 11 or above.

## Java Rules

[arl]: https://github.com/apple/apple_rules_lint


# contrib_rules_jvm

Handy rules for working with JVM-based projects in Bazel.

In order to use these in your own projects, in your `WORKSPACE` once
you've used an `http_archive`, you can load all the necessary
dependencies by:

```starlark
load("@contrib_rules_jvm//:repositories.bzl", "contrib_rules_jvm_deps")

contrib_rules_jvm_deps()

load("@contrib_rules_jvm//:setup.bzl", "contrib_rules_jvm_setup")

contrib_rules_jvm_setup()
```

If you're looking to get started quickly, then take a look at [java_test_suite](#java_test_suite) (a macro for generating a test suite from a `glob` of java test sources) and [java_junit5_test](#java_junit5_test) (a drop-in replacement for `java_test` that can run [JUnit5][junit5] tests)

## Linting

Many of the features in this repo are designed to be exposed via [apple_rules_lint][arl], which provides a framework for integrating linting checks into your builds. To take advantage of this perform the following steps:

```starlark
# In your WORKSPACE, after loading `apple_rules_lint`

load("@apple_rules_lint//lint:setup.bzl", "lint_setup")

lint_setup({
  # Note: this is an example config!
  "java-checkstyle": "@contrib_rules_jvm//java:checkstyle-default-config",
  "java-pmd": "@contrib_rules_jvm//java:pmd-config",
  "java-spotbugs": "@contrib_rules_jvm//java:spotbugs-default-config",
})
```

You are welcome to include all (or none!) of these rules, and linting
is "opt-in": if there's no `lint_setup` call in your repo's
`WORKSPACE` then everything will continue working just fine and no
additional lint tests will be generated.

The linters are configured using specific rules. The mappings are:

| Well known name | Lint config rule |
|-----------------|------------------|
| java-checkstyle | [checkstyle_config](#checkstyle_config) |
| java-pmd | [pmd_ruleset](#pmd_ruleset) |
| java-spotbugs | [spotbugs_config](#spotbugs_config) |

## Requirements

These rules require Java 11 or above.

## Java Rules

[arl]: https://github.com/apple/apple_rules_lint
[junit5]: https://junit.org/junit5/

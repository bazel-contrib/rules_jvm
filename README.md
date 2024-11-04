# contrib_rules_jvm

Handy rules for working with JVM-based projects in Bazel.

This ruleset is designed to complement `rules_java` (and Bazel's built-in Java rules), not replace them.

The intended way of working is that the standard `rules_java` rules are used, and this ruleset adds extra functionality which is compatible with `rules_java`.

## Using these rules

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

The gazelle plugin requires Go 1.18 or above.

## Java Rules

[arl]: https://github.com/apple/apple_rules_lint
[junit5]: https://junit.org/junit5/
<!-- Generated with Stardoc: http://skydoc.bazel.build -->



<a id="checkstyle_config"></a>

## checkstyle_config

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "checkstyle_config")

checkstyle_config(<a href="#checkstyle_config-name">name</a>, <a href="#checkstyle_config-data">data</a>, <a href="#checkstyle_config-checkstyle_binary">checkstyle_binary</a>, <a href="#checkstyle_config-config_file">config_file</a>, <a href="#checkstyle_config-output_format">output_format</a>)
</pre>

Rule allowing checkstyle to be configured. This is typically
used with the linting rules from `@apple_rules_lint` to configure how
checkstyle should run.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="checkstyle_config-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="checkstyle_config-data"></a>data |  Additional files to make available to Checkstyle such as any included XML files   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="checkstyle_config-checkstyle_binary"></a>checkstyle_binary |  Checkstyle binary to use.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `"@contrib_rules_jvm//java:checkstyle_cli"`  |
| <a id="checkstyle_config-config_file"></a>config_file |  The config file to use for all checkstyle tests   | <a href="https://bazel.build/concepts/labels">Label</a> | required |  |
| <a id="checkstyle_config-output_format"></a>output_format |  Output format to use. Defaults to plain   | String | optional |  `"plain"`  |


<a id="pmd_ruleset"></a>

## pmd_ruleset

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "pmd_ruleset")

pmd_ruleset(<a href="#pmd_ruleset-name">name</a>, <a href="#pmd_ruleset-format">format</a>, <a href="#pmd_ruleset-pmd_binary">pmd_binary</a>, <a href="#pmd_ruleset-rulesets">rulesets</a>, <a href="#pmd_ruleset-shallow">shallow</a>)
</pre>

Select a rule set for PMD tests.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="pmd_ruleset-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="pmd_ruleset-format"></a>format |  Generate report in the given format. One of html, text, or xml (default is xml)   | String | optional |  `"xml"`  |
| <a id="pmd_ruleset-pmd_binary"></a>pmd_binary |  PMD binary to use.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `"@contrib_rules_jvm//java:pmd"`  |
| <a id="pmd_ruleset-rulesets"></a>rulesets |  Use these rulesets.   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="pmd_ruleset-shallow"></a>shallow |  Use the targetted output to increase PMD's depth of processing   | Boolean | optional |  `True`  |


<a id="pmd_test"></a>

## pmd_test

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "pmd_test")

pmd_test(<a href="#pmd_test-name">name</a>, <a href="#pmd_test-srcs">srcs</a>, <a href="#pmd_test-ruleset">ruleset</a>, <a href="#pmd_test-target">target</a>)
</pre>

Use PMD to lint the `srcs`.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="pmd_test-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="pmd_test-srcs"></a>srcs |  -   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="pmd_test-ruleset"></a>ruleset |  -   | <a href="https://bazel.build/concepts/labels">Label</a> | required |  |
| <a id="pmd_test-target"></a>target |  -   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `None`  |


<a id="spotbugs_config"></a>

## spotbugs_config

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "spotbugs_config")

spotbugs_config(<a href="#spotbugs_config-name">name</a>, <a href="#spotbugs_config-effort">effort</a>, <a href="#spotbugs_config-exclude_filter">exclude_filter</a>, <a href="#spotbugs_config-fail_on_warning">fail_on_warning</a>, <a href="#spotbugs_config-plugin_list">plugin_list</a>, <a href="#spotbugs_config-spotbugs_binary">spotbugs_binary</a>)
</pre>

Configuration used for spotbugs, typically by the `//lint` rules.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="spotbugs_config-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="spotbugs_config-effort"></a>effort |  Effort can be min, less, default, more or max. Defaults to default   | String | optional |  `"default"`  |
| <a id="spotbugs_config-exclude_filter"></a>exclude_filter |  Report all bug instances except those matching the filter specified by this filter file   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `None`  |
| <a id="spotbugs_config-fail_on_warning"></a>fail_on_warning |  Whether to fail on warning, or just create a report. Defaults to True   | Boolean | optional |  `True`  |
| <a id="spotbugs_config-plugin_list"></a>plugin_list |  Specify a list of plugin Jar files to load   | <a href="https://bazel.build/concepts/labels">List of labels</a> | optional |  `[]`  |
| <a id="spotbugs_config-spotbugs_binary"></a>spotbugs_binary |  The spotbugs binary to run.   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `"@contrib_rules_jvm//java:spotbugs_cli"`  |


<a id="spotbugs_test"></a>

## spotbugs_test

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "spotbugs_test")

spotbugs_test(<a href="#spotbugs_test-name">name</a>, <a href="#spotbugs_test-deps">deps</a>, <a href="#spotbugs_test-config">config</a>, <a href="#spotbugs_test-only_output_jars">only_output_jars</a>)
</pre>

Use spotbugs to lint the `srcs`.

**ATTRIBUTES**


| Name  | Description | Type | Mandatory | Default |
| :------------- | :------------- | :------------- | :------------- | :------------- |
| <a id="spotbugs_test-name"></a>name |  A unique name for this target.   | <a href="https://bazel.build/concepts/labels#target-names">Name</a> | required |  |
| <a id="spotbugs_test-deps"></a>deps |  -   | <a href="https://bazel.build/concepts/labels">List of labels</a> | required |  |
| <a id="spotbugs_test-config"></a>config |  -   | <a href="https://bazel.build/concepts/labels">Label</a> | optional |  `"@contrib_rules_jvm//java:spotbugs-default-config"`  |
| <a id="spotbugs_test-only_output_jars"></a>only_output_jars |  If set to true, only the output jar of the target will be analyzed. Otherwise all transitive runtime dependencies will be analyzed   | Boolean | optional |  `True`  |


<a id="checkstyle_binary"></a>

## checkstyle_binary

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "checkstyle_binary")

checkstyle_binary(<a href="#checkstyle_binary-name">name</a>, <a href="#checkstyle_binary-main_class">main_class</a>, <a href="#checkstyle_binary-deps">deps</a>, <a href="#checkstyle_binary-runtime_deps">runtime_deps</a>, <a href="#checkstyle_binary-srcs">srcs</a>, <a href="#checkstyle_binary-visibility">visibility</a>, <a href="#checkstyle_binary-kwargs">kwargs</a>)
</pre>

Macro for quickly generating a `java_binary` target for use with `checkstyle_config`.

By default, this will set the `main_class` to point to the default one used by checkstyle
but it's ultimately a drop-replacement for straight `java_binary` target.

At least one of `runtime_deps`, `deps`, and `srcs` must be specified so that the
`java_binary` target will be valid.

An example would be:

```starlark
checkstyle_binary(
    name = "checkstyle_cli",
    runtime_deps = [
        artifact("com.puppycrawl.tools:checkstyle"),
    ]
)
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="checkstyle_binary-name"></a>name |  The name of the target   |  none |
| <a id="checkstyle_binary-main_class"></a>main_class |  The main class to use for checkstyle.   |  `"com.puppycrawl.tools.checkstyle.Main"` |
| <a id="checkstyle_binary-deps"></a>deps |  The deps required for compiling this binary. May be omitted.   |  `None` |
| <a id="checkstyle_binary-runtime_deps"></a>runtime_deps |  The deps required by checkstyle at runtime. May be omitted.   |  `None` |
| <a id="checkstyle_binary-srcs"></a>srcs |  If you're compiling your own `checkstyle` binary, the sources to use.   |  `None` |
| <a id="checkstyle_binary-visibility"></a>visibility |  <p align="center"> - </p>   |  `["//visibility:public"]` |
| <a id="checkstyle_binary-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


<a id="checkstyle_test"></a>

## checkstyle_test

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "checkstyle_test")

checkstyle_test(<a href="#checkstyle_test-name">name</a>, <a href="#checkstyle_test-size">size</a>, <a href="#checkstyle_test-timeout">timeout</a>, <a href="#checkstyle_test-kwargs">kwargs</a>)
</pre>



**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="checkstyle_test-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="checkstyle_test-size"></a>size |  <p align="center"> - </p>   |  `"medium"` |
| <a id="checkstyle_test-timeout"></a>timeout |  <p align="center"> - </p>   |  `"short"` |
| <a id="checkstyle_test-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


<a id="java_binary"></a>

## java_binary

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "java_binary")

java_binary(<a href="#java_binary-name">name</a>, <a href="#java_binary-kwargs">kwargs</a>)
</pre>

Adds linting tests to Bazel's own `java_binary`

**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="java_binary-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="java_binary-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


<a id="java_export"></a>

## java_export

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "java_export")

java_export(<a href="#java_export-name">name</a>, <a href="#java_export-maven_coordinates">maven_coordinates</a>, <a href="#java_export-pom_template">pom_template</a>, <a href="#java_export-deploy_env">deploy_env</a>, <a href="#java_export-visibility">visibility</a>, <a href="#java_export-kwargs">kwargs</a>)
</pre>

Adds linting tests to `rules_jvm_external`'s `java_export`

**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="java_export-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="java_export-maven_coordinates"></a>maven_coordinates |  <p align="center"> - </p>   |  none |
| <a id="java_export-pom_template"></a>pom_template |  <p align="center"> - </p>   |  `None` |
| <a id="java_export-deploy_env"></a>deploy_env |  <p align="center"> - </p>   |  `None` |
| <a id="java_export-visibility"></a>visibility |  <p align="center"> - </p>   |  `None` |
| <a id="java_export-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


<a id="java_junit5_test"></a>

## java_junit5_test

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "java_junit5_test")

java_junit5_test(<a href="#java_junit5_test-name">name</a>, <a href="#java_junit5_test-test_class">test_class</a>, <a href="#java_junit5_test-runtime_deps">runtime_deps</a>, <a href="#java_junit5_test-package_prefixes">package_prefixes</a>, <a href="#java_junit5_test-jvm_flags">jvm_flags</a>, <a href="#java_junit5_test-include_tags">include_tags</a>,
                 <a href="#java_junit5_test-exclude_tags">exclude_tags</a>, <a href="#java_junit5_test-include_engines">include_engines</a>, <a href="#java_junit5_test-exclude_engines">exclude_engines</a>, <a href="#java_junit5_test-kwargs">kwargs</a>)
</pre>

Run junit5 tests using Bazel.

This is designed to be a drop-in replacement for `java_test`, but
rather than using a JUnit4 runner it provides support for using
JUnit5 directly. The arguments are the same as used by `java_test`.

By default Bazel, and by extension this rule, assumes you want to
always run all of the tests in a class file.  The `include_tags`
and `exclude_tags` allows for selectively running specific tests
within a single class file based on your use of the `@Tag` Junit5
annotations. Please see [the JUnit 5
docs](https://junit.org/junit5/docs/current/user-guide/#running-tests-tags)
for more information about using JUnit5 tag annotation to control
test execution.

The generated target does not include any JUnit5 dependencies. If
you are using the standard `@maven` namespace for your
`maven_install` you can add these to your `deps` using
`JUNIT5_DEPS` or `JUNIT5_VINTAGE_DEPS` loaded from
`//java:defs.bzl`

**Note**: The junit5 runner prevents `System.exit` being called
using a `SecurityManager`, which means that one test can't
prematurely cause an entire test run to finish unexpectedly.
This security measure prohibits tests from setting their own `SecurityManager`.
To override this, set the `bazel.junit5runner.allowSettingSecurityManager` system property.

While the `SecurityManager` has been deprecated in recent Java
releases, there's no replacement yet. JEP 411 has this as one of
its goals, but this is not complete or available yet.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="java_junit5_test-name"></a>name |  The name of the test.   |  none |
| <a id="java_junit5_test-test_class"></a>test_class |  The Java class to be loaded by the test runner. If not specified, the class name will be inferred from a combination of the current bazel package and the `name` attribute.   |  `None` |
| <a id="java_junit5_test-runtime_deps"></a>runtime_deps |  <p align="center"> - </p>   |  `[]` |
| <a id="java_junit5_test-package_prefixes"></a>package_prefixes |  <p align="center"> - </p>   |  `[]` |
| <a id="java_junit5_test-jvm_flags"></a>jvm_flags |  <p align="center"> - </p>   |  `[]` |
| <a id="java_junit5_test-include_tags"></a>include_tags |  Junit5 tag expressions to include execution of tagged tests.   |  `[]` |
| <a id="java_junit5_test-exclude_tags"></a>exclude_tags |  Junit tag expressions to exclude execution of tagged tests.   |  `[]` |
| <a id="java_junit5_test-include_engines"></a>include_engines |  A list of JUnit Platform test engine IDs to include.   |  `[]` |
| <a id="java_junit5_test-exclude_engines"></a>exclude_engines |  A list of JUnit Platform test engine IDs to exclude.   |  `[]` |
| <a id="java_junit5_test-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


<a id="java_library"></a>

## java_library

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "java_library")

java_library(<a href="#java_library-name">name</a>, <a href="#java_library-kwargs">kwargs</a>)
</pre>

Adds linting tests to Bazel's own `java_library`

**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="java_library-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="java_library-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


<a id="java_test"></a>

## java_test

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "java_test")

java_test(<a href="#java_test-name">name</a>, <a href="#java_test-kwargs">kwargs</a>)
</pre>

Adds linting tests to Bazel's own `java_test`

**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="java_test-name"></a>name |  <p align="center"> - </p>   |  none |
| <a id="java_test-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


<a id="java_test_suite"></a>

## java_test_suite

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "java_test_suite")

java_test_suite(<a href="#java_test_suite-name">name</a>, <a href="#java_test_suite-srcs">srcs</a>, <a href="#java_test_suite-runner">runner</a>, <a href="#java_test_suite-test_suffixes">test_suffixes</a>, <a href="#java_test_suite-package">package</a>, <a href="#java_test_suite-deps">deps</a>, <a href="#java_test_suite-runtime_deps">runtime_deps</a>, <a href="#java_test_suite-size">size</a>, <a href="#java_test_suite-kwargs">kwargs</a>)
</pre>

Create a suite of java tests from `*Test.java` files.

This rule will create a `java_test` for each file which matches
any of the `test_suffixes` that are passed to this rule as
`srcs`. If any non-test sources are added these will first of all
be compiled into a `java_library` which will be added as a
dependency for each test target, allowing common utility functions
to be shared between tests.

The generated `java_test` targets will be named after the test file:
`FooTest.java` will create a `:FooTest` target.

In addition, a `test_suite` will be created, named using the `name`
attribute to allow all the tests to be run in one go.


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="java_test_suite-name"></a>name |  A unique name for this rule. Will be used to generate a `test_suite`   |  none |
| <a id="java_test_suite-srcs"></a>srcs |  Source files to create test rules for.   |  none |
| <a id="java_test_suite-runner"></a>runner |  One of `junit4` or `junit5`.   |  `"junit4"` |
| <a id="java_test_suite-test_suffixes"></a>test_suffixes |  The file name suffix used to identify if a file contains a test class.   |  `["Test.java"]` |
| <a id="java_test_suite-package"></a>package |  The package name used by the tests. If not set, this is inferred from the current bazel package name.   |  `None` |
| <a id="java_test_suite-deps"></a>deps |  A list of `java_*` dependencies.   |  `None` |
| <a id="java_test_suite-runtime_deps"></a>runtime_deps |  A list of `java_*` dependencies needed at runtime.   |  `[]` |
| <a id="java_test_suite-size"></a>size |  The size of the test, passed to `java_test`   |  `None` |
| <a id="java_test_suite-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


<a id="pmd_binary"></a>

## pmd_binary

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "pmd_binary")

pmd_binary(<a href="#pmd_binary-name">name</a>, <a href="#pmd_binary-main_class">main_class</a>, <a href="#pmd_binary-deps">deps</a>, <a href="#pmd_binary-runtime_deps">runtime_deps</a>, <a href="#pmd_binary-srcs">srcs</a>, <a href="#pmd_binary-visibility">visibility</a>, <a href="#pmd_binary-kwargs">kwargs</a>)
</pre>

Macro for quickly generating a `java_binary` target for use with `pmd_ruleset`.

By default, this will set the `main_class` to point to the default one used by PMD
but it's ultimately a drop-replacement for a regular `java_binary` target.

At least one of `runtime_deps`, `deps`, and `srcs` must be specified so that the
`java_binary` target will be valid.

An example would be:

```starlark
pmd_binary(
    name = "pmd",
    runtime_deps = [
        artifact("net.sourceforge.pmd:pmd-dist"),
    ],
)
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="pmd_binary-name"></a>name |  The name of the target   |  none |
| <a id="pmd_binary-main_class"></a>main_class |  The main class to use for PMD.   |  `"net.sourceforge.pmd.cli.PmdCli"` |
| <a id="pmd_binary-deps"></a>deps |  The deps required for compiling this binary. May be omitted.   |  `None` |
| <a id="pmd_binary-runtime_deps"></a>runtime_deps |  The deps required by PMD at runtime. May be omitted.   |  `None` |
| <a id="pmd_binary-srcs"></a>srcs |  If you're compiling your own PMD binary, the sources to use.   |  `None` |
| <a id="pmd_binary-visibility"></a>visibility |  <p align="center"> - </p>   |  `["//visibility:public"]` |
| <a id="pmd_binary-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


<a id="spotbugs_binary"></a>

## spotbugs_binary

<pre>
load("@contrib_rules_jvm//docs:stardoc-input.bzl", "spotbugs_binary")

spotbugs_binary(<a href="#spotbugs_binary-name">name</a>, <a href="#spotbugs_binary-main_class">main_class</a>, <a href="#spotbugs_binary-deps">deps</a>, <a href="#spotbugs_binary-runtime_deps">runtime_deps</a>, <a href="#spotbugs_binary-srcs">srcs</a>, <a href="#spotbugs_binary-visibility">visibility</a>, <a href="#spotbugs_binary-kwargs">kwargs</a>)
</pre>

Macro for quickly generating a `java_binary` target for use with `spotbugs_config`.

By default, this will set the `main_class` to point to the default one used by spotbugs
but it's ultimately a drop-replacement for a regular `java_binary` target.

At least one of `runtime_deps`, `deps`, and `srcs` must be specified so that the
`java_binary` target will be valid.

An example would be:

```starlark
spotbugs_binary(
    name = "spotbugs_cli",
    runtime_deps = [
        artifact("com.github.spotbugs:spotbugs"),
        artifact("org.slf4j:slf4j-jdk14"),
    ],
)
```


**PARAMETERS**


| Name  | Description | Default Value |
| :------------- | :------------- | :------------- |
| <a id="spotbugs_binary-name"></a>name |  The name of the target   |  none |
| <a id="spotbugs_binary-main_class"></a>main_class |  The main class to use for spotbugs.   |  `"edu.umd.cs.findbugs.LaunchAppropriateUI"` |
| <a id="spotbugs_binary-deps"></a>deps |  The deps required for compiling this binary. May be omitted.   |  `None` |
| <a id="spotbugs_binary-runtime_deps"></a>runtime_deps |  The deps required by spotbugs at runtime. May be omitted.   |  `None` |
| <a id="spotbugs_binary-srcs"></a>srcs |  If you're compiling your own `spotbugs` binary, the sources to use.   |  `None` |
| <a id="spotbugs_binary-visibility"></a>visibility |  <p align="center"> - </p>   |  `["//visibility:public"]` |
| <a id="spotbugs_binary-kwargs"></a>kwargs |  <p align="center"> - </p>   |  none |


## Freezing Dependencies

At runtime, a handful of dependencies are required by helper classes
in this project. Rather than pollute the default `@maven` workspace,
these are loaded into a `@contrib_rules_jvm_deps` workspace. These
dependencies are loaded using a call to `maven_install`, but we don't
want to force users to remember to load our own dependencies for
us. Instead, to add a new dependency to the project:

1. Update `contrib_rules_jvm_deps` in the `MODULE.bazel` file
2. Run `./tools/update-dependencies.sh`
3. Commit the updated files.

### Freezing your dependencies

As noted above, if you are building Bazel rules which require Java
parts, and hence Java dependencies, it can be useful to freeze these
dependencies. This process captures the list of dependencies into a
zip file.  This zip file is distributed with your Bazel rules. This
makes your rule more hermetic, your rule no longer relies on the user
to supply the correct dependencies, because they get resolved under
their own repository namespace rather than being intermingled with the
user's, and your dependencies no longer conflict with the users
selections. If you would like to create your dependencies as a frozen
file you need to do the following:

1. Create a maven install rule with your dependencies and with a
   unique name, this should be in or referenced by your WORKSPACE
   file.
   ```starlark
   maven_install(
        name = "frozen_deps",
        artifacts = [...],
        fail_if_repin_required = True,
        fetch_sources = True,
        maven_install_json = "@workspace//:frozen_deps_install.json",
   )
   
   load("@frozen_deps//:defs.bzl", "pinned_maven_install")
   
   pinned_maven_install()
   ```
2. Run `bazel run //tools:freeze-deps -- --repo <repo name> --zip
   <path/to/dependency.zip>`. The `<repo name>` matches the name used
   for the `maven_install()` rule above. This will pin the
   dependencies then collect them into the zip file.
3. Commit the zip file into your repo.
4. Add a `zip_repository()` rule to your WORKSPACE to configure the
   frozen dependencies:
   ```starlark
   maybe(
       zip_repository,
       name = "workspace_deps",
       path = "@workspace//path/to:dependency.zip",
   )
   ```
5. Make sure to pin the maven install from the repostitory:
   ```starlark
   load("@workspace_deps//:defs.bzl", "pinned_maven_install")
   
   pinned_maven_install()
   ```
6. Use the dependencies for your `java_library` rules from the frozen
   deps:
   ```starlark
   java_library(
        name = "my_library",
        srcs = glob(["*.java"]),
        deps = [
             "@workspace_deps//:my_frozen_dep",
        ],
   )
   ```

NOTE: If you need to generate the compat_repositories for the
dependencies, usually because your rule depends on another rule which
is still using the older compat repositories, you need to make the
following changes and abide by the following restrictions.

1. Add the `generate_compat_repositories = True,` attribute to the
   original `maven_install()` rule.
2. In step 2, add the parameter `--zip-repo workspace_deps` to match
   the name used in the `zip_repository()` rule (step 4). If you don't
   supply this, it uses the base name of the zip file, which may not
   be what you want.
3. In step 5, add the call to generate the compat_repositories:
   ```starlark
   load("@workspace_deps//:compat.bzl", "compat_repositories")
   
   compat_repositories()
   ```

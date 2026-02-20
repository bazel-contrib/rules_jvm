load("@rules_jvm_external//:defs.bzl", "DEFAULT_REPOSITORY_NAME", "artifact")
load("//java/private:library.bzl", "java_test")
load("//java/private:package.bzl", "get_package_name")

JUNIT5_RUNTIME_DEPS = [
    Label("//java/src/com/github/bazel_contrib/contrib_rules_jvm/junit5"),
]

_SYSTEM_EXIT_AGENT = Label("//java/src/com/github/bazel_contrib/contrib_rules_jvm/junit5/agent:system-exit-agent_deploy.jar")

def junit5_deps(repository_name = DEFAULT_REPOSITORY_NAME):
    return [
        artifact("org.junit.jupiter:junit-jupiter-engine", repository_name),
        artifact("org.junit.platform:junit-platform-launcher", repository_name),
        artifact("org.junit.platform:junit-platform-reporting", repository_name),
    ]

def junit5_vintage_deps(repository_name = DEFAULT_REPOSITORY_NAME):
    return junit5_deps(repository_name) + [
        artifact("org.junit.vintage:junit-vintage-engine", repository_name),
    ]

def junit5_jvm_flags(jvm_flags, include_tags = [], exclude_tags = [], include_engines = [], exclude_engines = [], report_generator = None):
    if include_tags:
        jvm_flags = jvm_flags + ["-DJUNIT5_INCLUDE_TAGS=" + ",".join(include_tags)]

    if exclude_tags:
        jvm_flags = jvm_flags + ["-DJUNIT5_EXCLUDE_TAGS=" + ",".join(exclude_tags)]

    if include_engines:
        jvm_flags = jvm_flags + ["-DJUNIT5_INCLUDE_ENGINES=%s" % ",".join(include_engines)]

    if exclude_engines:
        jvm_flags = jvm_flags + ["-DJUNIT5_EXCLUDE_ENGINES=%s" % ",".join(exclude_engines)]

    if report_generator:
        jvm_flags = jvm_flags + ["-DJUNIT5_REPORT_GENERATOR=%s" % report_generator]

    return jvm_flags

"""Dependencies typically required by JUnit 5 tests.

See `java_junit5_test` for more details.
"""
JUNIT5_DEPS = junit5_deps()

JUNIT5_VINTAGE_DEPS = junit5_vintage_deps()

def java_junit5_test(
        name,
        test_class = None,
        runtime_deps = [],
        package_prefixes = [],
        jvm_flags = [],
        include_tags = [],
        exclude_tags = [],
        include_engines = [],
        exclude_engines = [],
        report_generator = None,
        report_generator_deps = [],
        **kwargs):
    """Run junit5 tests using Bazel.

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
    using a Java agent, which means that one test can't prematurely
    cause an entire test run to finish unexpectedly.

    Args:
      name: The name of the test.
      test_class: The Java class to be loaded by the test runner. If not
        specified, the class name will be inferred from a combination of
        the current bazel package and the `name` attribute.
      include_tags: Junit5 tag expressions to include execution of tagged tests.
      exclude_tags: Junit tag expressions to exclude execution of tagged tests.
      include_engines: A list of JUnit Platform test engine IDs to include.
      exclude_engines: A list of JUnit Platform test engine IDs to exclude.
      report_generator: The name of an additional report generator to use.
        Can be 'legacy', 'open' or a class name.
      report_generator_deps: A list of labels to add to the runtime dependencies
        of the test, typically containing the report generator class.
    """
    if test_class:
        clazz = test_class
    else:
        clazz = get_package_name(package_prefixes) + name

    jvm_flags = junit5_jvm_flags(jvm_flags, include_tags, exclude_tags, include_engines, exclude_engines, report_generator)

    # Add the agent to prevent `System.exit` from being called
    jvm_flags = ["-javaagent:$(location %s)=$(location %s)" % (_SYSTEM_EXIT_AGENT, _SYSTEM_EXIT_AGENT)] + jvm_flags

    data = kwargs.pop("data", []) + [_SYSTEM_EXIT_AGENT]

    java_test(
        name = name,
        main_class = "com.github.bazel_contrib.contrib_rules_jvm.junit5.JUnit5Runner",
        test_class = clazz,
        runtime_deps = runtime_deps + JUNIT5_RUNTIME_DEPS + report_generator_deps,
        jvm_flags = jvm_flags,
        data = data,
        **kwargs
    )

load("//java/private:create_jvm_test_suite.bzl", "create_jvm_test_suite")
load("//java/private:java_test_suite_shared_constants.bzl", "DEFAULT_TEST_SUFFIXES")
load("//java/private:library.bzl", "java_library", "java_test")
load(":junit5.bzl", "java_junit5_test")

# Lifted from the Selenium project's work on migrating to Bazel.
# The key thing that this file adds is the ability to specify a
# suite of java tests by just globbing the test files.

def _define_library(name, **kwargs):
    java_library(
        name = name,
        **kwargs
    )

def _define_junit4_test(name, **kwargs):
    java_test(
        name = name,
        **kwargs
    )
    return name

def _define_junit5_test(name, **kwargs):
    java_junit5_test(
        name = name,
        include_engines = kwargs.pop("include_engines", None),
        exclude_engines = kwargs.pop("exclude_engines", None),
        **kwargs
    )
    return name

# Note: the keys in this match the keys in `create_jvm_test_suite.bzl`'s
# `_RUNNERS` constant
_TEST_GENERATORS = {
    "junit4": _define_junit4_test,
    "junit5": _define_junit5_test,
}

def java_test_suite(
        name,
        srcs,
        runner = "junit4",
        test_suffixes = DEFAULT_TEST_SUFFIXES,
        package = None,
        deps = None,
        runtime_deps = [],
        size = None,
        **kwargs):
    """Create a suite of java tests from `*Test.java` files.

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

    Args:
      name: A unique name for this rule. Will be used to generate a `test_suite`
      srcs: Source files to create test rules for.
      runner: One of `junit4` or `junit5`.
      package: The package name used by the tests. If not set, this is
        inferred from the current bazel package name.
      deps: A list of `java_*` dependencies.
      runtime_deps: A list of `java_*` dependencies needed at runtime.
      size: The size of the test, passed to `java_test`
      test_suffixes: The file name suffix used to identify if a file
        contains a test class.
    """
    create_jvm_test_suite(
        name,
        srcs = srcs,
        test_suffixes = test_suffixes,
        package = package,
        define_library = _define_library,
        # Default to bazel's default test runner if we don't know what people want
        define_test = _TEST_GENERATORS.get(runner, _define_junit4_test),
        runner = runner,
        deps = deps,
        runtime_deps = runtime_deps,
        size = size,
        **kwargs
    )

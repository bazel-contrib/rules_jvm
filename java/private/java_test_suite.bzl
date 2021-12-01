load("//java/private:create_jvm_test_suite.bzl", "create_jvm_test_suite")

# Lifted from the Selenium project's work on migrating to Bazel.
# The key thing that this file adds is the ability to specify a
# suite of java tests by just globbing the test files.

_LIBRARY_ATTRS = [
    "data",
    "javacopts",
    "plugins",
    "resources",
]

def _define_library(name, **kwargs):
    native.java_library(
        name = name,
        **kwargs
    )

def _define_test(name, **kwargs):
    native.java_test(
        name = name,
        **kwargs
    )

def java_test_suite(
        name,
        srcs,
        runner = "junit4",
        test_suffixes = ["Test.java"],
        deps = None,
        runtime_deps = [],
        tags = [],
        visibility = None,
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
        library_attributes = _LIBRARY_ATTRS,
        define_library = _define_library,
        define_test = _define_test,
        runner = runner,
        deps = deps,
        runtime_deps = runtime_deps,
        tags = tags,
        visibility = visibility,
        size = size,
        **kwargs
    )

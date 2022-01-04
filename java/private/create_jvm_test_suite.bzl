load("//java/private:package.bzl", "get_class_name")

_RUNNERS = {
    "junit4": {
        "main_class": None,
        "runtime_deps": [],
    },
    "junit5": {
        "main_class": "com.github.bazel_contrib.contrib_rules_jvm.junit5.JUnit5Runner",
        "runtime_deps": ["@contrib_rules_jvm//java:junit5-runner"],
    },
}

def _is_test(src, test_suffixes):
    for suffix in test_suffixes:
        if src.endswith(suffix):
            return True
    return False

def create_jvm_test_suite(
        name,
        srcs,
        test_suffixes,
        package,
        library_attributes,
        define_library,
        define_test,
        runner = "junit4",
        deps = None,
        runtime_deps = [],
        tags = [],
        visibility = None,
        size = None,
        **kwargs):
    """Generate a test suite for rules that "feel" like `java_test`.

    Given the list of `srcs`, this macro will generate:

      1. A `*_test` target per `src` that matches any of the `test_suffixes`
      2. A shared library that these tests depend on for any non-test `srcs`
      3. A `test_suite` tagged as `manual` that aggregates all the tests

    The reason for having a test target per test source file is to allow for
    better parallelization. Initial builds may be slower, but iterative builds
    while working with on unit tests should be faster, and this approach
    makes best use of the RBE.

    Args:
      name: The name of the generated test suite.
      srcs: A list of source files.
      test_suffixes: A list of suffixes (eg. `["Test.kt"]`)
      package: The package name to use. If `None`, a value will be
        calculated from the bazel package.
      library_attributes: Attributes to pass to `define_library`.
      define_library: A function that creates a `*_library` target.
      define_test: A function that creates a `*_test` target.
      runner: The junit runner to use. Either "junit4" or "junit5".
      deps: The list of dependencies to use when compiling.
      runtime_deps: The list of runtime deps to use when compiling.
      tags: Tags to use for generated targets.
      size: Bazel test size
    """

    if runner not in _RUNNERS:
        fail("Unknown java_test_suite runner. Must be one of {}".format(_RUNNERS.keys()))
    runner_params = _RUNNERS.get(runner)

    # First, grab any interesting attrs
    library_attrs = {attr: kwargs[attr] for attr in library_attributes if attr in kwargs}

    test_srcs = [src for src in srcs if _is_test(src, test_suffixes)]
    nontest_srcs = [src for src in srcs if not _is_test(src, test_suffixes)]

    if nontest_srcs:
        # Build a shared test library to use for everything. If we don't do this,
        # each rule needs to compile all sources, and that seems grossly inefficient.
        # Only include the non-test sources since we don't want all tests to re-run
        # when only one test source changes.
        define_library(
            name = "%s-test-lib" % name,
            deps = deps,
            srcs = nontest_srcs,
            testonly = True,
            **library_attrs
        )
        deps.append(":%s-test-lib" % name)

    tests = []

    for src in test_srcs:
        suffix = src.rfind(".")
        test_name = src[:suffix]
        tests.append(test_name)
        test_class = get_class_name(package, src)

        define_test(
            name = test_name,
            size = size,
            srcs = [src],
            test_class = test_class,
            main_class = runner_params["main_class"],
            deps = deps,
            tags = tags,
            runtime_deps = runtime_deps + [dep for dep in runner_params["runtime_deps"] if dep not in runtime_deps],
            visibility = ["//visibility:private"],
            **kwargs
        )

    native.test_suite(
        name = name,
        tests = tests,
        tags = ["manual"] + tags,
        visibility = visibility,
    )

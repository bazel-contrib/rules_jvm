load("@apple_rules_lint//lint:defs.bzl", "get_lint_config")
load("@rules_jvm_external//:defs.bzl", _java_export = "java_export")
load("//java/private:checkstyle.bzl", "checkstyle_test")

def _create_lint_tests(name, **kwargs):
    srcs = kwargs.get("srcs", [])

    if len(srcs) == 0:
        return

    tags = kwargs.get("tags", [])

    checkstyle = get_lint_config("java-checkstyle", tags)
    if checkstyle != None:
        checkstyle_test(
            name = "%s-checkstyle" % name,
            srcs = srcs,
            config = checkstyle,
            # Do not keep the parent tags: we typically want to run lint tests
            # regardless of the library or test tags (e.g. even if we exclude
            # sidecar tests, we want to lint them).
            tags = ["lint", "checkstyle", "java-checkstyle"],
            size = "small",
            timeout = "moderate",
        )

def java_binary(name, **kwargs):
    """Adds linting tests to Bazel's own `java_binary`"""
    _create_lint_tests(name, **kwargs)
    native.java_binary(name = name, **kwargs)

def java_library(name, **kwargs):
    """Adds linting tests to Bazel's own `java_library`"""
    _create_lint_tests(name, **kwargs)
    native.java_library(name = name, **kwargs)

def java_test(name, **kwargs):
    """Adds linting tests to Bazel's own `java_test`"""
    _create_lint_tests(name, **kwargs)
    native.java_test(name = name, **kwargs)

def java_export(name, maven_coordinates, pom_template = None, deploy_env = None, visibility = None, **kwargs):
    """Adds linting tests to `rules_jvm_external`'s `java_export`"""
    _create_lint_tests(name, **kwargs)
    _java_export(
        name = name,
        maven_coordinates = maven_coordinates,
        pom_template = pom_template,
        deploy_env = deploy_env,
        visibility = visibility,
        **kwargs
    )

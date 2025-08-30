load("@apple_rules_lint//lint:defs.bzl", "get_lint_config")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("@rules_java//java:java_binary.bzl", _java_binary = "java_binary")
load("@rules_java//java:java_library.bzl", _java_library = "java_library")
load("@rules_java//java:java_test.bzl", _java_test = "java_test")
load("@rules_jvm_external//:defs.bzl", _java_export = "java_export")
load("//java/private:checkstyle.bzl", "checkstyle_test")
load("//java/private:pmd.bzl", "pmd_test")
load("//java/private:spotbugs.bzl", "spotbugs_test")

def create_lint_tests(name, **kwargs):
    srcs = kwargs.get("srcs", [])

    if len(srcs) == 0:
        return

    tags = kwargs.get("tags", [])

    checkstyle = get_lint_config("java-checkstyle", tags)
    if checkstyle != None:
        maybe(
            checkstyle_test,
            name = "%s-checkstyle" % name,
            srcs = srcs,
            config = checkstyle,
            # Do not keep the parent tags: we typically want to run lint tests
            # regardless of the library or test tags (e.g. even if we exclude
            # sidecar tests, we want to lint them).
            tags = ["lint", "checkstyle", "java-checkstyle"],
        )

    pmd = get_lint_config("java-pmd", tags)
    if pmd != None:
        maybe(
            pmd_test,
            name = "%s-pmd" % name,
            srcs = srcs,
            target = ":%s" % name,
            ruleset = pmd,
            tags = ["lint", "pmd", "java-pmd"],
            size = "small",
            timeout = "moderate",
        )

    spotbugs = get_lint_config("java-spotbugs", tags)
    if spotbugs != None:
        maybe(
            spotbugs_test,
            name = "%s-spotbugs" % name,
            config = spotbugs,
            only_output_jars = True,
            deps = [
                ":%s" % name,
            ],
            tags = ["lint", "spotbugs", "java-spotbugs"],
            size = "small",
            timeout = "moderate",
        )

def java_binary(name, **kwargs):
    """Adds linting tests to Bazel's own `java_binary`"""
    create_lint_tests(name, **kwargs)
    _java_binary(name = name, **kwargs)

def java_library(name, **kwargs):
    """Adds linting tests to Bazel's own `java_library`"""
    create_lint_tests(name, **kwargs)
    _java_library(name = name, **kwargs)

def java_test(name, **kwargs):
    """Adds linting tests to Bazel's own `java_test`"""
    create_lint_tests(name, **kwargs)
    _java_test(name = name, **kwargs)

def java_export(name, maven_coordinates, pom_template = None, deploy_env = None, visibility = None, **kwargs):
    """Adds linting tests to `rules_jvm_external`'s `java_export`"""

    # Only run the lint tests on the java_library target and not the java_export target which will include
    # other sources within it.
    # https://github.com/bazel-contrib/rules_jvm_external/blob/534e62d14655e01048b90ef89ab1acaf0caa7348/private/rules/java_export.bzl#L85C5-L85C31
    lib_name = "%s-lib" % name
    create_lint_tests(lib_name, **kwargs)
    _java_export(
        name = name,
        maven_coordinates = maven_coordinates,
        pom_template = pom_template,
        deploy_env = deploy_env,
        visibility = visibility,
        **kwargs
    )

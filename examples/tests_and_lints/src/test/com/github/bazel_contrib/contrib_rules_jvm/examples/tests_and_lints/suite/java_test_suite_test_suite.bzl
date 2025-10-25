load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load("@contrib_rules_jvm//java:defs.bzl", "java_junit5_test", "java_test_suite")
load("@rules_java//java:java_test.bzl", "java_test")

TargetInfo = provider(
    doc = "Information relating to the target under test.",
    fields = ["attr"],
)

def _target_info_aspect_impl(target, ctx):
    return TargetInfo(
        attr = ctx.rule.attr,
    )

target_info_aspect = aspect(
    implementation = _target_info_aspect_impl,
)

def _check_standard_test_suite_tags_test_impl(ctx):
    env = analysistest.begin(ctx)
    target_under_test = analysistest.target_under_test(env)

    asserts.false(env, ("custom" in target_under_test[TargetInfo].attr.tags))

    return analysistest.end(env)

check_standard_test_suite_tags_test = analysistest.make(
    _check_standard_test_suite_tags_test_impl,
    extra_target_under_test_aspects = [target_info_aspect],
)

def _check_custom_test_suite_tags_test_impl(ctx):
    env = analysistest.begin(ctx)
    target_under_test = analysistest.target_under_test(env)

    asserts.true(env, ("custom" in target_under_test[TargetInfo].attr.tags))

    return analysistest.end(env)

check_custom_test_suite_tags_test = analysistest.make(
    _check_custom_test_suite_tags_test_impl,
    extra_target_under_test_aspects = [target_info_aspect],
)

def _custom_junit4_test(name, **kwargs):
    kwargs["tags"] = ["manual", "custom"]
    java_test(name = name, **kwargs)

    return name

def _custom_junit5_test(name, **kwargs):
    kwargs["tags"] = ["manual", "custom"]
    java_junit5_test(name = name, **kwargs)

    return name

def java_test_suite_test_suite(name):
    java_test_suite(
        name = "StandardJunit4Suite",
        tags = ["manual"],
        srcs = ["StandardJunit4SuiteTest.java"],
        test_suffixes = ["Test.java"],
        deps = [],
        runner = "junit4",
    )

    java_test_suite(
        name = "StandardJunit5Suite",
        tags = ["manual"],
        srcs = ["StandardJunit5SuiteTest.java"],
        test_suffixes = ["Test.java"],
        deps = [],
        runner = "junit5",
    )

    java_test_suite(
        name = "CustomJunit4Suite",
        tags = ["manual"],
        srcs = ["CustomJunit4SuiteTest.java"],
        test_suffixes = ["Test.java"],
        deps = [],
        runner = "junit4",
        test_generators = {
            "junit4": _custom_junit4_test,
        },
    )

    java_test_suite(
        name = "CustomJunit5Suite",
        tags = ["manual"],
        srcs = ["CustomJunit5SuiteTest.java"],
        test_suffixes = ["Test.java"],
        deps = [],
        runner = "junit5",
        test_generators = {
            "junit5": _custom_junit5_test,
        },
    )

    check_standard_test_suite_tags_test(
        name = "standard_junit4_suite_runner_test",
        target_under_test = ":StandardJunit4SuiteTest",
    )

    check_standard_test_suite_tags_test(
        name = "standard_junit5_suite_runner_test",
        target_under_test = ":StandardJunit5SuiteTest",
    )

    check_custom_test_suite_tags_test(
        name = "custom_junit4_suite_runner_test",
        target_under_test = ":CustomJunit4SuiteTest",
    )

    check_custom_test_suite_tags_test(
        name = "custom_junit5_suite_runner_test",
        target_under_test = ":CustomJunit5SuiteTest",
    )

    native.test_suite(
        name = name,
        tests = [
            ":standard_junit4_suite_runner_test",
            ":standard_junit5_suite_runner_test",
            ":custom_junit4_suite_runner_test",
            ":custom_junit5_suite_runner_test",
        ],
    )

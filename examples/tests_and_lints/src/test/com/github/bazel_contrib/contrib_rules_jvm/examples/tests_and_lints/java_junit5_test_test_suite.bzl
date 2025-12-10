load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")
load("@contrib_rules_jvm//java:defs.bzl", "java_junit5_test", "junit5_jvm_flags")
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

def _attr_string_value_test_impl(ctx):
    env = analysistest.begin(ctx)
    target_under_test = analysistest.target_under_test(env)

    asserts.equals(env, ctx.attr.check_value, getattr(target_under_test[TargetInfo].attr, ctx.attr.check_name))

    return analysistest.end(env)

attr_string_value_test = analysistest.make(
    _attr_string_value_test_impl,
    extra_target_under_test_aspects = [target_info_aspect],
    attrs = {
        "check_name": attr.string(mandatory = True),
        "check_value": attr.string(mandatory = True),
    },
)

def _attr_string_list_value_test_impl(ctx):
    env = analysistest.begin(ctx)
    target_under_test = analysistest.target_under_test(env)

    asserts.equals(env, ctx.attr.check_value, getattr(target_under_test[TargetInfo].attr, ctx.attr.check_name))

    return analysistest.end(env)

attr_string_list_value_test = analysistest.make(
    _attr_string_list_value_test_impl,
    extra_target_under_test_aspects = [target_info_aspect],
    attrs = {
        "check_name": attr.string(mandatory = True),
        "check_value": attr.string_list(mandatory = True),
    },
)

def custom_junit5_test(name, **kwargs):
    jvm_flags = junit5_jvm_flags(
        jvm_flags = kwargs.pop("jvm_flags", []),
        include_tags = kwargs.pop("include_tags", []),
        exclude_tags = kwargs.pop("exclude_tags", []),
        include_engines = kwargs.pop("include_engines", []),
        exclude_engines = kwargs.pop("exclude_engines", []),
    )

    java_test(
        name = name,
        main_class = "com.example.CustomMainClass",
        jvm_flags = jvm_flags,
        **kwargs
    )

def java_junit5_test_test_suite(name):
    java_junit5_test(
        name = "StandardMainClassTest",
        tags = ["manual"],
    )

    custom_junit5_test(
        name = "CustomMainClassTest",
        include_tags = ["include_junit5_test"],
        tags = ["manual"],
    )

    attr_string_value_test(
        name = "custom_main_class_test",
        target_under_test = ":CustomMainClassTest",
        check_name = "main_class",
        check_value = "com.example.CustomMainClass",
    )

    attr_string_list_value_test(
        name = "custom_jvm_flags_test",
        target_under_test = ":CustomMainClassTest",
        check_name = "jvm_flags",
        check_value = ["-DJUNIT5_INCLUDE_TAGS=include_junit5_test"],
    )

    attr_string_value_test(
        name = "standard_main_class_test",
        target_under_test = ":StandardMainClassTest",
        check_name = "main_class",
        check_value = "com.github.bazel_contrib.contrib_rules_jvm.junit5.JUnit5Runner",
    )

    native.test_suite(
        name = name,
        tests = [
            ":custom_main_class_test",
            ":standard_main_class_test",
        ],
    )

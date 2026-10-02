load("@bazel_skylib//lib:unittest.bzl", "analysistest", "asserts")

_TargetInfo = provider(fields = ["attr"])

def _target_info_aspect_impl(target, ctx):
    return _TargetInfo(attr = ctx.rule.attr)

_target_info_aspect = aspect(implementation = _target_info_aspect_impl)

def _target_names(targets):
    return [target.label.name for target in targets]

def _overridden_test_has_merged_attributes_impl(ctx):
    env = analysistest.begin(ctx)
    attr = analysistest.target_under_test(env)[_TargetInfo].attr

    asserts.equals(env, "large", attr.size)
    asserts.equals(env, 2, attr.shard_count)
    asserts.equals(env, ["suite-tag", "test-tag"], attr.tags)
    asserts.equals(env, ["Helper.java", "DefaultTest.java"], _target_names(attr.data))
    asserts.equals(env, {
        "DEP": "$(location :make-var-dep)",
        "SHARED": "test",
        "SUITE_ONLY": "suite",
        "TEST_ONLY": "test",
    }, attr.env)
    asserts.equals(env, ["SUITE_ENV", "TEST_ENV"], attr.env_inherit)
    asserts.equals(env, [
        "-Dsuite=true",
        "-Dtest=true",
    ], attr.jvm_flags)
    asserts.true(env, "make-var-dep" in _target_names(attr.deps))

    return analysistest.end(env)

_overridden_test_has_merged_attributes_test = analysistest.make(
    _overridden_test_has_merged_attributes_impl,
    extra_target_under_test_aspects = [_target_info_aspect],
)

def _default_test_keeps_suite_attributes_impl(ctx):
    env = analysistest.begin(ctx)
    attr = analysistest.target_under_test(env)[_TargetInfo].attr

    asserts.equals(env, "small", attr.size)
    asserts.equals(env, ["suite-tag"], attr.tags)
    asserts.equals(env, ["Helper.java"], _target_names(attr.data))
    asserts.equals(env, {
        "SHARED": "suite",
        "SUITE_ONLY": "suite",
    }, attr.env)
    asserts.equals(env, ["SUITE_ENV"], attr.env_inherit)
    asserts.equals(env, ["-Dsuite=true"], attr.jvm_flags)

    return analysistest.end(env)

_default_test_keeps_suite_attributes_test = analysistest.make(
    _default_test_keeps_suite_attributes_impl,
    extra_target_under_test_aspects = [_target_info_aspect],
)

def _selected_env_is_merged_impl(ctx):
    env = analysistest.begin(ctx)
    attr = analysistest.target_under_test(env)[_TargetInfo].attr

    asserts.equals(env, {
        "SELECTED": "test",
        "SHARED": "test",
        "SUITE_ONLY": "suite",
    }, attr.env)

    return analysistest.end(env)

_selected_env_is_merged_test = analysistest.make(
    _selected_env_is_merged_impl,
    extra_target_under_test_aspects = [_target_info_aspect],
)

def _additional_library_srcs_augment_the_library_impl(ctx):
    env = analysistest.begin(ctx)
    attr = analysistest.target_under_test(env)[_TargetInfo].attr

    asserts.equals(env, ["Helper.java", "SharedTest.java"], _target_names(attr.srcs))

    return analysistest.end(env)

_additional_library_srcs_augment_the_library_test = analysistest.make(
    _additional_library_srcs_augment_the_library_impl,
    extra_target_under_test_aspects = [_target_info_aspect],
)

def _aggregate_suite_keeps_only_suite_tags_impl(ctx):
    env = analysistest.begin(ctx)
    attr = analysistest.target_under_test(env)[_TargetInfo].attr

    asserts.equals(env, ["manual", "suite-tag"], attr.tags)

    return analysistest.end(env)

_aggregate_suite_keeps_only_suite_tags_test = analysistest.make(
    _aggregate_suite_keeps_only_suite_tags_impl,
    extra_target_under_test_aspects = [_target_info_aspect],
)

def test_suite_overrides_test_suite(name):
    _overridden_test_has_merged_attributes_test(
        name = "overridden_test_has_merged_attributes",
        target_under_test = ":OverriddenTest",
    )

    _default_test_keeps_suite_attributes_test(
        name = "default_test_keeps_suite_attributes",
        target_under_test = ":DefaultTest",
    )

    _selected_env_is_merged_test(
        name = "selected_env_is_merged",
        target_under_test = ":SharedTest",
    )

    _additional_library_srcs_augment_the_library_test(
        name = "additional_library_srcs_augment_the_library",
        target_under_test = ":fixture-test-lib",
    )

    _aggregate_suite_keeps_only_suite_tags_test(
        name = "aggregate_suite_keeps_only_suite_tags",
        target_under_test = ":fixture",
    )

    native.test_suite(
        name = name,
        tests = [
            ":additional_library_srcs_augment_the_library",
            ":aggregate_suite_keeps_only_suite_tags",
            ":default_test_keeps_suite_attributes",
            ":overridden_test_has_merged_attributes",
            ":selected_env_is_merged",
        ],
    )

load("@contrib_rules_jvm//java:defs.bzl", "java_test_suite")

java_test_suite(
    name = "withhelpers-tests",
    srcs = [
        "Helper.java",
        "NotWorldWithHelpersTest.java",
        "withdirectory/AnotherTest.java",
        "withdirectory/Helper.java",
    ],
    runner = "junit5",
    runtime_deps = [
        "@maven//:org_junit_jupiter_junit_jupiter_engine",
        "@maven//:org_junit_platform_junit_platform_launcher",
        "@maven//:org_junit_platform_junit_platform_reporting",
    ],
    deps = ["@maven//:junit_junit"],
)

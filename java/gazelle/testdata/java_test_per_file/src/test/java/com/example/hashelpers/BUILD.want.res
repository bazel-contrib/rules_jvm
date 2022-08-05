load("@rules_java//java:defs.bzl", "java_library", "java_test")

java_library(
    name = "hashelpers",
    testonly = True,
    srcs = ["Helper.java"],
    deps = [
        "@maven//:junit_junit",
        "@maven//:org_hamcrest_hamcrest_all",
    ],
)

java_test(
    name = "AppTest",
    srcs = ["AppTest.java"],
    test_class = "com.example.hashelpers.AppTest",
    deps = [
        "@maven//:junit_junit",
        "@maven//:org_hamcrest_hamcrest_all",
    ],
)

java_test(
    name = "OtherAppTest",
    srcs = ["OtherAppTest.java"],
    test_class = "com.example.hashelpers.OtherAppTest",
    deps = [
        "@maven//:junit_junit",
        "@maven//:org_hamcrest_hamcrest_all",
    ],
)

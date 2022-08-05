load("@rules_java//java:defs.bzl", "java_library", "java_test")

java_library(
    name = "modulewithhelpers",
    testonly = True,
    srcs = [
        "Helper.java",
        "withdirectory/Helper.java",
    ],
    deps = [
        "@maven//:junit_junit",
        "@maven//:org_hamcrest_hamcrest_all",
    ],
)

java_test(
    name = "com_example_modulewithhelpers_AppTest",
    srcs = ["AppTest.java"],
    test_class = "com.example.modulewithhelpers.AppTest",
    deps = [
        "@maven//:junit_junit",
        "@maven//:org_hamcrest_hamcrest_all",
    ],
)

java_test(
    name = "com_example_modulewithhelpers_withdirectory_AnotherTest",
    srcs = ["withdirectory/AnotherTest.java"],
    test_class = "com.example.modulewithhelpers.withdirectory.AnotherTest",
    deps = [
        "@maven//:junit_junit",
        "@maven//:org_hamcrest_hamcrest_all",
    ],
)

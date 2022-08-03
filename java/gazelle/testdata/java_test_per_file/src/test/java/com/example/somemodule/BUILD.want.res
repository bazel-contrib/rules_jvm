load("@rules_java//java:defs.bzl", "java_test")

java_test(
    name = "com_example_somemodule_withdirectory_ATest",
    srcs = ["withdirectory/ATest.java"],
    test_class = "com.example.somemodule.withdirectory.ATest",
    deps = [
        "@maven//:junit_junit",
        "@maven//:org_hamcrest_hamcrest_all",
    ],
)

java_test(
    name = "com_example_somemodule_withdirectory_AnotherTest",
    srcs = ["withdirectory/AnotherTest.java"],
    test_class = "com.example.somemodule.withdirectory.AnotherTest",
    deps = [
        "@maven//:junit_junit",
        "@maven//:org_hamcrest_hamcrest_all",
    ],
)

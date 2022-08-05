load("@rules_java//java:defs.bzl", "java_binary", "java_library")
load("@contrib_rules_jvm//java:defs.bzl", "java_test_suite")

java_library(
    name = "module1",
    srcs = [
        "src/main/java/com/example/hello/Hello.java",
        "src/main/java/com/example/hello/world/World.java",
    ],
    visibility = ["//:__subpackages__"],
)

java_binary(
    name = "Hello",
    main_class = "com.example.hello.Hello",
    visibility = ["//visibility:public"],
    runtime_deps = [":module1"],
)

java_test_suite(
    name = "module1-tests",
    srcs = [
        "src/test/java/com/example/hello/world/OtherWorldTest.java",
        "src/test/java/com/example/hello/world/WorldTest.java",
    ],
    runner = "junit5",
    runtime_deps = [
        "@maven//:org_junit_jupiter_junit_jupiter_engine",
        "@maven//:org_junit_platform_junit_platform_launcher",
        "@maven//:org_junit_platform_junit_platform_reporting",
    ],
    deps = ["@maven//:junit_junit"],
)

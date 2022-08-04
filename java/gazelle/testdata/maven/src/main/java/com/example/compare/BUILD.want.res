load("@rules_java//java:defs.bzl", "java_binary", "java_library")

java_library(
    name = "compare",
    srcs = ["Compare.java"],
    visibility = ["//:__subpackages__"],
    deps = ["@maven//:com_google_guava_guava"],
)

java_binary(
    name = "Compare",
    main_class = "com.example.compare.Compare",
    visibility = ["//visibility:public"],
    runtime_deps = [":compare"],
)

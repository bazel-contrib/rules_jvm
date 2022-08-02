load("@rules_java//java:defs.bzl", "java_library")

java_library(
    name = "library",
    srcs = ["Library.java"],
    visibility = ["//:__subpackages__"],
    deps = ["@maven//:com_google_guava_guava"],
)

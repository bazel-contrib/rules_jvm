load("@rules_java//java:defs.bzl", "java_library")

java_library(
    name = "example2",
    srcs = ["Main.java"],
    visibility = ["//:__subpackages__"],
)

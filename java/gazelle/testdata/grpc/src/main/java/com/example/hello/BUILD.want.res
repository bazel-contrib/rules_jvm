load("@rules_java//java:defs.bzl", "java_binary", "java_library")

java_library(
    name = "hello",
    srcs = ["hello.java"],
    visibility = ["//:__subpackages__"],
)

java_binary(
    name = "Hello",
    main_class = "com.example.hello.Hello",
    visibility = ["//visibility:public"],
    runtime_deps = [":hello"],
)

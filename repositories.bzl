load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("//java/private:zip_repository.bzl", "zip_repository")

def contrib_rules_jvm_deps():
    maybe(
        http_archive,
        name = "apple_rules_lint",
        sha256 = "8feab4b08a958b10cb2abb7f516652cd770b582b36af6477884b3bba1f2f0726",
        strip_prefix = "apple_rules_lint-0.1.1",
        url = "https://github.com/apple/apple_rules_lint/archive/0.1.1.zip",
    )
    maybe(
        http_archive,
        name = "io_bazel_stardoc",
        sha256 = "c9794dcc8026a30ff67cf7cf91ebe245ca294b20b071845d12c192afe243ad72",
        url = "https://github.com/bazelbuild/stardoc/releases/download/0.5.0/stardoc-0.5.0.tar.gz",
    )

    maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "07b4117379dde7ab382345c3b0f5edfc6b7cff6c93756eac63da121e0bbcc5de",
        strip_prefix = "bazel-skylib-1.1.1",
        url = "https://github.com/bazelbuild/bazel-skylib/archive/1.1.1.tar.gz",
    )

    maybe(
        http_archive,
        name = "com_google_protobuf",
        sha256 = "c6003e1d2e7fefa78a3039f19f383b4f3a61e81be8c19356f85b6461998ad3db",
        strip_prefix = "protobuf-3.17.3",
        urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.17.3.tar.gz"],
    )

    maybe(
        zip_repository,
        name = "contrib_rules_jvm_deps",
        path = "@contrib_rules_jvm//java/private:contrib_rules_jvm_deps.zip",
    )

    maybe(
        http_archive,
        name = "io_grpc_grpc_java",
        sha256 = "e3781bcab2a410a7cd138f13b2e6a643e111575f6811b154c570f4d020e87507",
        strip_prefix = "grpc-java-1.44.0",
        urls = ["https://github.com/grpc/grpc-java/archive/v1.44.0.tar.gz"],
    )

    maybe(
        http_archive,
        name = "rules_jvm_external",
        sha256 = "cd1a77b7b02e8e008439ca76fd34f5b07aecb8c752961f9640dea15e9e5ba1ca",
        strip_prefix = "rules_jvm_external-4.2",
        url = "https://github.com/bazelbuild/rules_jvm_external/archive/4.2.zip",
        patches = [
            "@contrib_rules_jvm//java/private:make-docs-visible.patch",
        ],
        patch_args = ["-p1"],
    )

def contrib_rules_jvm_gazelle_deps():
    maybe(
        http_archive,
        name = "bazel_gazelle",
        sha256 = "dae13a7c6adb742174aafd340ebcb36016de231bd4f926f79c140c7d9b599fb0",
        strip_prefix = "bazel-gazelle-757e291d1befe9174fb1fcf0d9ade733cbb6b904",
        urls = [
            "https://github.com/bazelbuild/bazel-gazelle/archive/757e291d1befe9174fb1fcf0d9ade733cbb6b904.zip",
        ],
    )

    maybe(
        http_archive,
        name = "io_bazel_rules_go",
        sha256 = "f2dcd210c7095febe54b804bb1cd3a58fe8435a909db2ec04e31542631cf715c",
        urls = [
            "https://github.com/bazelbuild/rules_go/releases/download/v0.31.0/rules_go-v0.31.0.zip",
        ],
    )

    maybe(
        http_archive,
        name = "rules_proto",
        sha256 = "66bfdf8782796239d3875d37e7de19b1d94301e8972b3cbd2446b332429b4df1",
        strip_prefix = "rules_proto-4.0.0",
        url = "https://github.com/bazelbuild/rules_proto/archive/refs/tags/4.0.0.tar.gz",
    )

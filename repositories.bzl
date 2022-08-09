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
        sha256 = "05fb57bb4ad68a360470420a3b6f5317e4f722839abc5b17ec4ef8ed465aaa47",
        url = "https://github.com/bazelbuild/stardoc/releases/download/0.5.2/stardoc-0.5.2.tar.gz",
    )

    maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "710c2ca4b4d46250cdce2bf8f5aa76ea1f0cba514ab368f2988f70e864cfaf51",
        strip_prefix = "bazel-skylib-1.2.1",
        url = "https://github.com/bazelbuild/bazel-skylib/archive/1.2.1.tar.gz",
    )

    maybe(
        http_archive,
        name = "com_google_protobuf",
        sha256 = "d7d204a59fd0d2d2387bd362c2155289d5060f32122c4d1d922041b61191d522",
        strip_prefix = "protobuf-3.21.5",
        urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.21.5.tar.gz"],
    )

    maybe(
        zip_repository,
        name = "contrib_rules_jvm_deps",
        path = "@contrib_rules_jvm//java/private:contrib_rules_jvm_deps.zip",
    )

    maybe(
        http_archive,
        name = "io_grpc_grpc_java",
        sha256 = "88b12b2b4e0beb849eddde98d5373f2f932513229dbf9ec86cc8e4912fc75e79",
        strip_prefix = "grpc-java-1.48.1",
        urls = ["https://github.com/grpc/grpc-java/archive/v1.48.1.tar.gz"],
    )

    maybe(
        http_archive,
        name = "rules_jvm_external",
        sha256 = "f3945075856f44fdbdaece3872842c6950a4f07146b40b0cd01d36480dfa5436",
        strip_prefix = "rules_jvm_external-3957fdc382b5a404fccdde74b91c1e614e07e6bd",
        url = "https://github.com/bazelbuild/rules_jvm_external/archive/3957fdc382b5a404fccdde74b91c1e614e07e6bd.zip",
    )

def contrib_rules_jvm_gazelle_deps():
    maybe(
        http_archive,
        name = "bazel_gazelle",
        sha256 = "501deb3d5695ab658e82f6f6f549ba681ea3ca2a5fb7911154b5aa45596183fa",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.26.0/bazel-gazelle-v0.26.0.tar.gz",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.26.0/bazel-gazelle-v0.26.0.tar.gz",
        ],
    )

    maybe(
        http_archive,
        name = "com_github_bazelbuild_buildtools",
        sha256 = "e3bb0dc8b0274ea1aca75f1f8c0c835adbe589708ea89bf698069d0790701ea3",
        strip_prefix = "buildtools-5.1.0",
        url = "https://github.com/bazelbuild/buildtools/archive/5.1.0.tar.gz",
    )

    maybe(
        http_archive,
        name = "io_bazel_rules_go",
        sha256 = "16e9fca53ed6bd4ff4ad76facc9b7b651a89db1689a2877d6fd7b82aa824e366",
        urls = [
            "https://github.com/bazelbuild/rules_go/releases/download/v0.34.0/rules_go-v0.34.0.zip",
        ],
    )

    maybe(
        http_archive,
        name = "io_grpc_grpc_java",
        sha256 = "88b12b2b4e0beb849eddde98d5373f2f932513229dbf9ec86cc8e4912fc75e79",
        strip_prefix = "grpc-java-1.48.1",
        urls = ["https://github.com/grpc/grpc-java/archive/v1.48.1.tar.gz"],
    )

    maybe(
        http_archive,
        name = "rules_proto",
        sha256 = "e017528fd1c91c5a33f15493e3a398181a9e821a804eb7ff5acdd1d2d6c2b18d",
        strip_prefix = "rules_proto-4.0.0-3.20.0",
        urls = [
            "https://github.com/bazelbuild/rules_proto/archive/refs/tags/4.0.0-3.20.0.tar.gz",
        ],
    )

load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("//java/private:zip_repository.bzl", "zip_repository")

def contrib_rules_jvm_deps():
    maybe(
        http_archive,
        name = "apple_rules_lint",
        sha256 = "78223d7c8372d025fc377a17aca83961c3d9d4829bcbb9f12476cea212ea70ea",
        strip_prefix = "apple_rules_lint-0.2.0",
        url = "https://github.com/apple/apple_rules_lint/archive/0.2.0.zip",
    )
    maybe(
        http_archive,
        name = "io_bazel_stardoc",
        sha256 = "3fd8fec4ddec3c670bd810904e2e33170bedfe12f90adf943508184be458c8bb",
        url = "https://github.com/bazelbuild/stardoc/releases/download/0.5.3/stardoc-0.5.3.tar.gz",
    )

    maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "3b620033ca48fcd6f5ef2ac85e0f6ec5639605fa2f627968490e52fc91a9932f",
        strip_prefix = "bazel-skylib-1.3.0",
        url = "https://github.com/bazelbuild/bazel-skylib/archive/1.3.0.tar.gz",
    )

    maybe(
        http_archive,
        name = "com_google_protobuf",
        sha256 = "ce2fbea3c78147a41b2a922485d283137845303e5e1b6cbd7ece94b96ade7031",
        strip_prefix = "protobuf-3.21.7",
        urls = ["https://github.com/protocolbuffers/protobuf/archive/v3.21.7.tar.gz"],
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
        sha256 = "d1dcd42aa9afee37f2e00697ae348d586b72f8e40ee79ee55acb57cc257064ad",
        strip_prefix = "rules_jvm_external-7b439d1e1751d44916530a19806e86a2a105c03b",
        url = "https://github.com/bazelbuild/rules_jvm_external/archive/7b439d1e1751d44916530a19806e86a2a105c03b.zip",
    )

def contrib_rules_jvm_gazelle_deps():
    maybe(
        http_archive,
        name = "bazel_gazelle",
        sha256 = "efbbba6ac1a4fd342d5122cbdfdb82aeb2cf2862e35022c752eaddffada7c3f3",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.27.0/bazel-gazelle-v0.27.0.tar.gz",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.27.0/bazel-gazelle-v0.27.0.tar.gz",
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

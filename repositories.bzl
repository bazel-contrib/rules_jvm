load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("//java/private:zip_repository.bzl", "zip_repository")
load("//third_party:protobuf_version.bzl", "PROTOBUF_VERSION")

def contrib_rules_jvm_deps():
    # We need the latest version of `rules_license`, but many of our deps pull in an older version
    maybe(
        http_archive,
        name = "rules_license",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_license/releases/download/1.0.0/rules_license-1.0.0.tar.gz",
            "https://github.com/bazelbuild/rules_license/releases/download/1.0.0/rules_license-1.0.0.tar.gz",
        ],
        sha256 = "26d4021f6898e23b82ef953078389dd49ac2b5618ac564ade4ef87cced147b38",
    )

    maybe(
        http_archive,
        name = "apple_rules_lint",
        strip_prefix = "apple_rules_lint-0.4.0",
        sha256 = "483ea03d73d5fb33275d029da8d36811243fc32dfa4dc73a43acbb6f4b1af621",
        url = "https://github.com/apple/apple_rules_lint/releases/download/0.4.0/apple_rules_lint-0.4.0.tar.gz",
    )

    maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "bc283cdfcd526a52c3201279cda4bc298652efa898b10b4db0837dc51652756f",
        url = "https://github.com/bazelbuild/bazel-skylib/releases/download/1.7.1/bazel-skylib-1.7.1.tar.gz",
    )

    maybe(
        http_archive,
        name = "com_google_protobuf",
        sha256 = "75be42bd736f4df6d702a0e4e4d30de9ee40eac024c4b845d17ae4cc831fe4ae",
        strip_prefix = "protobuf-{}".format(PROTOBUF_VERSION),
        urls = ["https://github.com/protocolbuffers/protobuf/archive/v{}.tar.gz".format(PROTOBUF_VERSION)],
    )

    # The `rules_java` major version is tied to the major version of Bazel that it supports,
    # so this is different from the version in the MODULE file.
    major_version = native.bazel_version.partition(".")[0]
    if major_version == "5":
        maybe(
            http_archive,
            name = "rules_java",
            urls = [
                "https://github.com/bazelbuild/rules_java/releases/download/5.5.1/rules_java-5.5.1.tar.gz",
            ],
            sha256 = "73b88f34dc251bce7bc6c472eb386a6c2b312ed5b473c81fe46855c248f792e0",
        )

    elif major_version == "6":
        maybe(
            http_archive,
            name = "rules_java",
            urls = [
                "https://github.com/bazelbuild/rules_java/releases/download/6.5.2/rules_java-6.5.2.tar.gz",
            ],
            sha256 = "16bc94b1a3c64f2c36ceecddc9e09a643e80937076b97e934b96a8f715ed1eaa",
        )

    else:
        maybe(
            http_archive,
            name = "rules_java",
            urls = [
                "https://github.com/bazelbuild/rules_java/releases/download/7.12.1/rules_java-7.12.1.tar.gz",
            ],
            sha256 = "dfbadbb37a79eb9e1cc1e156ecb8f817edf3899b28bc02410a6c1eb88b1a6862",
        )

    maybe(
        zip_repository,
        name = "contrib_rules_jvm_deps",
        path = "@contrib_rules_jvm//java/private:contrib_rules_jvm_deps.zip",
    )

    # This is required by `rules_jvm_external`
    maybe(
        http_archive,
        name = "bazel_features",
        sha256 = "b4b145c19e08fd48337f53c383db46398d0a810002907ff0c590762d926e05be",
        strip_prefix = "bazel_features-1.18.0",
        url = "https://github.com/bazel-contrib/bazel_features/releases/download/v1.18.0/bazel_features-v1.18.0.tar.gz",
    )

    maybe(
        http_archive,
        name = "rules_jvm_external",
        sha256 = "85776be6d8fe64abf26f463a8e12cd4c15be927348397180a01693610da7ec90",
        strip_prefix = "rules_jvm_external-6.4",
        url = "https://github.com/bazel-contrib/rules_jvm_external/releases/download/6.4/rules_jvm_external-6.4.tar.gz",
    )

def contrib_rules_jvm_gazelle_deps():
    io_grpc_grpc_java()

    http_archive(
        name = "googleapis",
        sha256 = "9d1a930e767c93c825398b8f8692eca3fe353b9aaadedfbcf1fca2282c85df88",
        strip_prefix = "googleapis-64926d52febbf298cb82a8f472ade4a3969ba922",
        urls = [
            "https://github.com/googleapis/googleapis/archive/64926d52febbf298cb82a8f472ade4a3969ba922.zip",
        ],
    )

    maybe(
        http_archive,
        name = "bazel_gazelle",
        sha256 = "b760f7fe75173886007f7c2e616a21241208f3d90e8657dc65d36a771e916b6a",
        url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.39.1/bazel-gazelle-v0.39.1.tar.gz",
    )

    maybe(
        http_archive,
        name = "com_github_bazelbuild_buildtools",
        sha256 = "061472b3e8b589fb42233f0b48798d00cf9dee203bd39502bd294e6b050bc6c2",
        strip_prefix = "buildtools-7.1.0",
        url = "https://github.com/bazelbuild/buildtools/archive/refs/tags/v7.1.0.tar.gz",
    )

    maybe(
        http_archive,
        name = "io_bazel_rules_go",
        sha256 = "f4a9314518ca6acfa16cc4ab43b0b8ce1e4ea64b81c38d8a3772883f153346b8",
        url = "https://github.com/bazelbuild/rules_go/releases/download/v0.50.1/rules_go-v0.50.1.zip",
    )

    maybe(
        http_archive,
        name = "rules_proto",
        sha256 = "6fb6767d1bef535310547e03247f7518b03487740c11b6c6adb7952033fe1295",
        strip_prefix = "rules_proto-6.0.2",
        url = "https://github.com/bazelbuild/rules_proto/releases/download/6.0.2/rules_proto-6.0.2.tar.gz",
    )

    # We need to expand the contents of `@rules_proto//proto:repositories.bzl" here so
    # we can continue the two-step initialisation process
    maybe(
        http_archive,
        name = "rules_cc",
        sha256 = "4aeb102efbcfad509857d7cb9c5456731e8ce566bfbf2960286a2ec236796cc3",
        strip_prefix = "rules_cc-2f8c04c04462ab83c545ab14c0da68c3b4c96191",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_cc/archive/2f8c04c04462ab83c545ab14c0da68c3b4c96191.tar.gz",
            "https://github.com/bazelbuild/rules_cc/archive/2f8c04c04462ab83c545ab14c0da68c3b4c96191.tar.gz",
        ],
    )

    maybe(
        http_archive,
        name = "rules_python",
        sha256 = "d70cd72a7a4880f0000a6346253414825c19cdd40a28289bdf67b8e6480edff8",
        strip_prefix = "rules_python-0.28.0",
        url = "https://github.com/bazelbuild/rules_python/releases/download/0.28.0/rules_python-0.28.0.tar.gz",
    )

    # And other repos we need, apparently

    maybe(
        http_archive,
        name = "bazel_features",
        sha256 = "3646ffd447753490b77d2380fa63f4d55dd9722e565d84dfda01536b48e183da",
        strip_prefix = "bazel_features-1.19.0",
        url = "https://github.com/bazel-contrib/bazel_features/releases/download/v1.19.0/bazel_features-v1.19.0.tar.gz",
    )

    maybe(
        http_archive,
        name = "rules_pkg",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_pkg/releases/download/1.0.1/rules_pkg-1.0.1.tar.gz",
            "https://github.com/bazelbuild/rules_pkg/releases/download/1.0.1/rules_pkg-1.0.1.tar.gz",
        ],
        sha256 = "d20c951960ed77cb7b341c2a59488534e494d5ad1d30c4818c736d57772a9fef",
    )

def io_grpc_grpc_java():
    maybe(
        http_archive,
        name = "io_grpc_grpc_java",
        sha256 = "17dd91014032a147c978ae99582fddd950f5444388eae700cf51eda0326ad2f9",
        strip_prefix = "grpc-java-1.56.1",
        urls = ["https://github.com/grpc/grpc-java/archive/v1.56.1.tar.gz"],
    )

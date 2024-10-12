load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("//java/private:zip_repository.bzl", "zip_repository")
load("//third_party:protobuf_version.bzl", "PROTOBUF_VERSION")

def contrib_rules_jvm_deps():
    maybe(
        http_archive,
        name = "apple_rules_lint",
        sha256 = "7c3cc45a95e3ef6fbc484a4234789a027e11519f454df63cbb963ac499f103f9",
        strip_prefix = "apple_rules_lint-0.3.2",
        url = "https://github.com/apple/apple_rules_lint/archive/refs/tags/0.3.2.tar.gz",
    )

    maybe(
        http_archive,
        name = "io_bazel_stardoc",
        sha256 = "dfbc364aaec143df5e6c52faf1f1166775a5b4408243f445f44b661cfdc3134f",
        url = "https://github.com/bazelbuild/stardoc/releases/download/0.5.6/stardoc-0.5.6.tar.gz",
    )

    maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "66ffd9315665bfaafc96b52278f57c7e2dd09f5ede279ea6d39b2be471e7e3aa",
        url = "https://github.com/bazelbuild/bazel-skylib/releases/download/1.4.2/bazel-skylib-1.4.2.tar.gz",
    )

    maybe(
        http_archive,
        name = "bazel_skylib_gazelle_plugin",
        sha256 = "3327005dbc9e49cc39602fb46572525984f7119a9c6ffe5ed69fbe23db7c1560",
        url = "https://github.com/bazelbuild/bazel-skylib/releases/download/1.4.2/bazel-skylib-gazelle-plugin-1.4.2.tar.gz",
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
                "https://github.com/bazelbuild/rules_java/releases/download/7.3.2/rules_java-7.3.2.tar.gz",
            ],
            sha256 = "3121a00588b1581bd7c1f9b550599629e5adcc11ba9c65f482bbd5cfe47fdf30",
        )

    maybe(
        zip_repository,
        name = "contrib_rules_jvm_deps",
        path = "@contrib_rules_jvm//java/private:contrib_rules_jvm_deps.zip",
    )

    maybe(
        http_archive,
        name = "rules_jvm_external",
        sha256 = "808cb5c30b5f70d12a2a745a29edc46728fd35fa195c1762a596b63ae9cebe05",
        strip_prefix = "rules_jvm_external-6.2",
        url = "https://github.com/bazelbuild/rules_jvm_external/releases/download/6.2/rules_jvm_external-6.2.tar.gz",
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

    # We need https://github.com/bazelbuild/bazel-gazelle/pull/1798
    maybe(
        http_archive,
        name = "bazel_gazelle",
        sha256 = "d1d72a9abd6dee362354274fa9b60ced8f50ee1f10f9b9fef90b4acfb98d477b",
        strip_prefix = "bazel-gazelle-ba2ce367a545e0bdd74a7abca40ef5e0a0cb8dcb",
        url = "https://github.com/bazelbuild/bazel-gazelle/archive/ba2ce367a545e0bdd74a7abca40ef5e0a0cb8dcb.zip",
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
        sha256 = "d6ab6b57e48c09523e93050f13698f708428cfd5e619252e369d377af6597707",
        url = "https://github.com/bazelbuild/rules_go/releases/download/v0.43.0/rules_go-v0.43.0.zip",
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
        name = "rules_license",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_license/releases/download/0.0.7/rules_license-0.0.7.tar.gz",
            "https://github.com/bazelbuild/rules_license/releases/download/0.0.7/rules_license-0.0.7.tar.gz",
        ],
        sha256 = "4531deccb913639c30e5c7512a054d5d875698daeb75d8cf90f284375fe7c360",
    )

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

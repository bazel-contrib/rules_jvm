load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("//third_party:protobuf_version.bzl", "PROTOBUF_VERSION")
load("//java/private:zip_repository.bzl", "zip_repository")

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

    maybe(
        zip_repository,
        name = "contrib_rules_jvm_deps",
        path = "@contrib_rules_jvm//java/private:contrib_rules_jvm_deps.zip",
    )

    maybe(
        http_archive,
        name = "rules_jvm_external",
        sha256 = "d31e369b854322ca5098ea12c69d7175ded971435e55c18dd9dd5f29cc5249ac",
        strip_prefix = "rules_jvm_external-5.3",
        url = "https://github.com/bazelbuild/rules_jvm_external/releases/download/5.3/rules_jvm_external-5.3.tar.gz",
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

    # We need https://github.com/bazelbuild/bazel-gazelle/pull/1686
    maybe(
        http_archive,
        name = "bazel_gazelle",
        sha256 = "2b4aa25b5288c6498cbb10cd68b11c87f3f114b279c551263ed964b8cc23d016",
        strip_prefix = "bazel-gazelle-8e09916b20b52a0346e77a3203fc0b50825fedcd",
        url = "https://github.com/bazelbuild/bazel-gazelle/archive/8e09916b20b52a0346e77a3203fc0b50825fedcd.tar.gz",
    )

    maybe(
        http_archive,
        name = "com_github_bazelbuild_buildtools",
        sha256 = "05c3c3602d25aeda1e9dbc91d3b66e624c1f9fdadf273e5480b489e744ca7269",
        strip_prefix = "buildtools-6.4.0",
        url = "https://github.com/bazelbuild/buildtools/archive/refs/tags/v6.4.0.tar.gz",
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
        sha256 = "dc3fb206a2cb3441b485eb1e423165b231235a1ea9b031b4433cf7bc1fa460dd",
        # We intentionally format it with the protobuf version to ensure that,
        # if we try to upgrade protobuf without upgrading rules_proto, it crashes.
        strip_prefix = "rules_proto-5.3.0-{}".format(PROTOBUF_VERSION),
        url = "https://github.com/bazelbuild/rules_proto/archive/refs/tags/5.3.0-{}.tar.gz".format(PROTOBUF_VERSION),
    )

def io_grpc_grpc_java():
    maybe(
        http_archive,
        name = "io_grpc_grpc_java",
        sha256 = "17dd91014032a147c978ae99582fddd950f5444388eae700cf51eda0326ad2f9",
        strip_prefix = "grpc-java-1.56.1",
        urls = ["https://github.com/grpc/grpc-java/archive/v1.56.1.tar.gz"],
    )

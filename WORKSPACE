workspace(name = "contrib_rules_jvm")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "apple_rules_lint",
    sha256 = "8feab4b08a958b10cb2abb7f516652cd770b582b36af6477884b3bba1f2f0726",
    strip_prefix = "apple_rules_lint-0.1.1",
    url = "https://github.com/apple/apple_rules_lint/archive/0.1.1.zip",
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "cb1f888a5363f89945ece2254576bd84730aac2642d64ee1bbcfb500735a8beb",
    strip_prefix = "bazel-gazelle-2190265f2fabd2383765fdfbed158106ebe81e9b",
    url = "https://github.com/bazelbuild/bazel-gazelle/archive/2190265f2fabd2383765fdfbed158106ebe81e9b.tar.gz",
)

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "2b1641428dff9018f9e85c0384f03ec6c10660d935b750e3fa1492a281a53b0f",
    url = "https://github.com/bazelbuild/rules_go/releases/download/v0.29.0/rules_go-v0.29.0.zip",
)

http_archive(
    name = "io_bazel_stardoc",
    sha256 = "c9794dcc8026a30ff67cf7cf91ebe245ca294b20b071845d12c192afe243ad72",
    url = "https://github.com/bazelbuild/stardoc/releases/download/0.5.0/stardoc-0.5.0.tar.gz",
)

http_archive(
    name = "rules_proto",
    sha256 = "66bfdf8782796239d3875d37e7de19b1d94301e8972b3cbd2446b332429b4df1",
    strip_prefix = "rules_proto-4.0.0",
    url = "https://github.com/bazelbuild/rules_proto/archive/refs/tags/4.0.0.tar.gz",
)

http_archive(
    name = "rules_python",
    sha256 = "cdf6b84084aad8f10bf20b46b77cb48d83c319ebe6458a18e9d2cebf57807cdd",
    strip_prefix = "rules_python-0.8.1",
    url = "https://github.com/bazelbuild/rules_python/archive/refs/tags/0.8.1.tar.gz",
)

load("@apple_rules_lint//lint:repositories.bzl", "lint_deps")

lint_deps()

load("@apple_rules_lint//lint:setup.bzl", "lint_setup")

lint_setup({
    "java-checkstyle": "//java:checkstyle-default-config",
    "java-pmd": "//java:pmd-config",
    "java-spotbugs": "//java:spotbugs-default-config",
})

load("//:repositories.bzl", "contrib_rules_jvm_deps")

contrib_rules_jvm_deps()

load("//:setup.bzl", "contrib_rules_jvm_setup")

# gazelle:repository_macro third_party/go/repositories.bzl%go_deps
contrib_rules_jvm_setup()

load("@rules_jvm_external//:defs.bzl", "maven_install")
load("@io_grpc_grpc_java//:repositories.bzl", "IO_GRPC_GRPC_JAVA_ARTIFACTS")

# This only exists to give us a target to use with `./tools/update-dependencies.sh`.
# If you update this, then please re-run that script and commit the changes to repo.
maven_install(
    name = "frozen_deps",
    artifacts = [
        "com.github.javaparser:javaparser-core:3.22.1",
        "com.github.javaparser:javaparser-symbol-solver-core:3.22.1",
        "com.google.code.findbugs:jsr305:3.0.2",
        "com.google.guava:guava:30.1.1-jre",
        "commons-cli:commons-cli:1.5.0",
        "io.grpc:grpc-api:1.40.0",
        "io.grpc:grpc-core:1.40.0",
        "io.grpc:grpc-netty:1.40.0",
        "io.grpc:grpc-stub:1.40.0",
        "org.slf4j:slf4j-simple:1.7.32",

        # These can be versioned independently of the versions in `repositories.bzl`
        # so long as the version numbers are higher.
        "org.junit.jupiter:junit-jupiter-engine:5.8.1",
        "org.junit.jupiter:junit-jupiter-api:5.8.1",
        "org.junit.platform:junit-platform-launcher:1.8.1",
        "org.junit.platform:junit-platform-reporting:1.8.1",
        "org.junit.vintage:junit-vintage-engine:5.8.1",

        # Checkstyle deps
        "com.puppycrawl.tools:checkstyle:9.2",

        # PMD deps
        "net.sourceforge.pmd:pmd-dist:6.41.0",

        # Spotbugs deps
        # We don't want to force people to use 1.8-beta
        # but we can't use the `maven` macros because
        # we've not loaded rules yet. Fortunately, the
        # expansion is easy :)
        {
            "group": "com.github.spotbugs",
            "artifact": "spotbugs",
            "version": "4.5.3",
            "exclusions": [
                {
                    "group": "org.slf4j",
                    "artifact": "slf4j-api",
                },
            ],
        },
        "org.slf4j:slf4j-api:1.7.32",
        "org.slf4j:slf4j-jdk14:1.7.32",
    ] + IO_GRPC_GRPC_JAVA_ARTIFACTS,
    fail_if_repin_required = True,
    fetch_sources = True,
    generate_compat_repositories = True,
    maven_install_json = "@contrib_rules_jvm//:frozen_deps_install.json",
    repositories = [
        "https://repo1.maven.org/maven2",
    ],
)

load("@frozen_deps//:defs.bzl", frozen_deps_pmi = "pinned_maven_install")

frozen_deps_pmi()

# These are used for our own tests.
maven_install(
    artifacts = [
        "com.google.code.findbugs:annotations:3.0.1",
        "com.google.googlejavaformat:google-java-format:1.15.0",

        # These can be versioned independently of the versions in `repositories.bzl`
        # so long as the version numbers are higher.
        "org.junit.jupiter:junit-jupiter-engine:5.8.2",
        "org.junit.jupiter:junit-jupiter-api:5.8.2",
        "org.junit.platform:junit-platform-launcher:1.8.2",
        "org.junit.platform:junit-platform-reporting:1.8.2",
        "org.junit.vintage:junit-vintage-engine:5.8.2",
    ],
    fail_if_repin_required = True,
    fetch_sources = True,
    maven_install_json = "@//:maven_install.json",
    repositories = [
        "https://repo1.maven.org/maven2",
    ],
)

load("@maven//:defs.bzl", maven_pmi = "pinned_maven_install")

maven_pmi()

load("@io_bazel_stardoc//:setup.bzl", "stardoc_repositories")

stardoc_repositories()

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.18")

gazelle_dependencies()

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")

rules_proto_dependencies()

rules_proto_toolchains()

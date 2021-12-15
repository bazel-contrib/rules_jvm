workspace(name = "contrib_rules_jvm")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "apple_rules_lint",
    sha256 = "8feab4b08a958b10cb2abb7f516652cd770b582b36af6477884b3bba1f2f0726",
    strip_prefix = "apple_rules_lint-0.1.1",
    url = "https://github.com/apple/apple_rules_lint/archive/0.1.1.zip",
)

load("@apple_rules_lint//lint:repositories.bzl", "lint_deps")

lint_deps()

load("@apple_rules_lint//lint:setup.bzl", "lint_setup")

lint_setup({
    "java-checkstyle": "//java:checkstyle-default-config",
    "java-pmd": "//:pmd-config",
    "java-spotbugs": "//java:spotbugs-default-config",
})

load("//:repositories.bzl", "contrib_rules_jvm_deps")

contrib_rules_jvm_deps()

load("//:setup.bzl", "contrib_rules_jvm_setup")

contrib_rules_jvm_setup()

load("@rules_jvm_external//:defs.bzl", "maven_install")

# This only exists to give us a target to use with `//bin:freeze-deps.py` If
# you update this, then please re-run that script and commit the changes to
# repo
maven_install(
    name = "frozen_deps",
    artifacts = [
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
            "version": "4.5.0",
            "exclusions": [
                {
                    "group": "org.slf4j",
                    "artifact": "slf4j-api",
                },
            ],
        },
        "org.slf4j:slf4j-api:1.7.32",
        "org.slf4j:slf4j-jdk14:1.7.32",
    ],
    fail_if_repin_required = True,
    fetch_sources = True,
    maven_install_json = "@contrib_rules_jvm//:frozen_deps_install.json",
    repositories = [
        "https://repo1.maven.org/maven2",
    ],
)

load("@frozen_deps//:defs.bzl", "pinned_maven_install")

pinned_maven_install()

# These are used for our own tests.
maven_install(
    artifacts = [
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

load("@maven//:defs.bzl", "pinned_maven_install")

pinned_maven_install()

# Dependencies for generating `bzl_library` targets using Gazelle
http_archive(
    name = "io_bazel_rules_go",
    sha256 = "2b1641428dff9018f9e85c0384f03ec6c10660d935b750e3fa1492a281a53b0f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.29.0/rules_go-v0.29.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.29.0/rules_go-v0.29.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "de69a09dc70417580aabf20a28619bb3ef60d038470c7cf8442fafcf627c21cb",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.24.0/bazel-gazelle-v0.24.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.24.0/bazel-gazelle-v0.24.0.tar.gz",
    ],
)

http_archive(
    name = "io_bazel_stardoc",
    sha256 = "c9794dcc8026a30ff67cf7cf91ebe245ca294b20b071845d12c192afe243ad72",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/stardoc/releases/download/0.5.0/stardoc-0.5.0.tar.gz",
        "https://github.com/bazelbuild/stardoc/releases/download/0.5.0/stardoc-0.5.0.tar.gz",
    ],
)

load("@io_bazel_stardoc//:setup.bzl", "stardoc_repositories")

stardoc_repositories()

############################################
# Gazelle, for generating bzl_library targets
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.17.2")

gazelle_dependencies()

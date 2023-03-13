workspace(name = "contrib_rules_jvm")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "rules_python",
    sha256 = "497ca47374f48c8b067d786b512ac10a276211810f4a580178ee9b9ad139323a",
    strip_prefix = "rules_python-0.16.1",
    url = "https://github.com/bazelbuild/rules_python/archive/refs/tags/0.16.1.tar.gz",
)

load("//:repositories.bzl", "contrib_rules_jvm_deps", "contrib_rules_jvm_gazelle_deps")

contrib_rules_jvm_deps()

contrib_rules_jvm_gazelle_deps()

load("@apple_rules_lint//lint:repositories.bzl", "lint_deps")

lint_deps()

load("@apple_rules_lint//lint:setup.bzl", "lint_setup")

lint_setup({
    "java-checkstyle": "//java:checkstyle-default-config",
    "java-pmd": "//java:pmd-config",
    "java-spotbugs": "//java:spotbugs-default-config",
})

load("//:setup.bzl", "contrib_rules_jvm_setup")

# gazelle:repository_macro third_party/go/repositories.bzl%go_deps
contrib_rules_jvm_setup()

load("//:gazelle_setup.bzl", "contrib_rules_jvm_gazelle_setup")

contrib_rules_jvm_gazelle_setup()

load("@rules_jvm_external//:defs.bzl", "maven_install")
load("@io_grpc_grpc_java//:repositories.bzl", "IO_GRPC_GRPC_JAVA_ARTIFACTS")

# This only exists to give us a target to use with `./tools/update-dependencies.sh`.
# If you update this, then please re-run that script and commit the changes to repo.
maven_install(
    name = "frozen_deps",
    artifacts = [
        "com.google.code.findbugs:jsr305:3.0.2",
        "com.google.errorprone:error_prone_annotations:2.11.0",
        "com.google.guava:guava:30.1.1-jre",
        "commons-cli:commons-cli:1.5.0",
        "io.grpc:grpc-api:1.40.0",
        "io.grpc:grpc-core:1.40.0",
        "io.grpc:grpc-netty:1.40.0",
        "io.grpc:grpc-services:1.40.0",
        "io.grpc:grpc-stub:1.40.0",
        "org.slf4j:slf4j-simple:1.7.32",
        "com.google.googlejavaformat:google-java-format:1.15.0",

        # These can be versioned independently of the versions in `repositories.bzl`
        # so long as the version numbers are higher.
        "org.junit.jupiter:junit-jupiter-engine:5.8.1",
        "org.junit.jupiter:junit-jupiter-api:5.8.1",
        "org.junit.platform:junit-platform-launcher:1.8.1",
        "org.junit.platform:junit-platform-reporting:1.8.1",
        "org.junit.vintage:junit-vintage-engine:5.8.1",

        # Open Test Alliance for the JVM dep
        "org.opentest4j:opentest4j:1.2.0",

        # Checkstyle deps
        "com.puppycrawl.tools:checkstyle:10.2",

        # PMD deps
        "net.sourceforge.pmd:pmd-dist:6.46.0",

        # Spotbugs deps
        # We don't want to force people to use 1.8-beta
        # but we can't use the `maven` macros because
        # we've not loaded rules yet. Fortunately, the
        # expansion is easy :)
        {
            "group": "com.github.spotbugs",
            "artifact": "spotbugs",
            "version": "4.7.0",
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
    name = "contrib_rules_jvm_tests",
    artifacts = [
        # These can be versioned independently of the versions in `repositories.bzl`
        # so long as the version numbers are higher.
        "org.junit.jupiter:junit-jupiter-engine:5.8.2",
        "org.junit.jupiter:junit-jupiter-api:5.8.2",
        "org.junit.jupiter:junit-jupiter-params:5.8.2",
        "org.junit.platform:junit-platform-launcher:1.8.2",
        "org.junit.platform:junit-platform-reporting:1.8.2",
        "org.junit.vintage:junit-vintage-engine:5.8.2",
        "org.mockito:mockito-core:4.8.1",
    ],
    fail_if_repin_required = True,
    fetch_sources = True,
    maven_install_json = "@//:contrib_rules_jvm_tests_install.json",
    repositories = [
        "https://repo1.maven.org/maven2",
    ],
)

load("@contrib_rules_jvm_tests//:defs.bzl", maven_pmi = "pinned_maven_install")

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

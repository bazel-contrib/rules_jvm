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

load("@io_grpc_grpc_java//:repositories.bzl", "IO_GRPC_GRPC_JAVA_ARTIFACTS")
load("@rules_jvm_external//:defs.bzl", "maven_install")
load("//third_party:protobuf_version.bzl", "PROTOBUF_JAVA_VERSION")

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
        "org.junit.platform:junit-platform-suite:1.8.2",
        "org.junit.platform:junit-platform-suite-api:1.8.2",
        "org.junit.platform:junit-platform-suite-engine:1.8.2",
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

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains(version = "1.21.5")

gazelle_dependencies()

load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")

rules_proto_dependencies()

rules_proto_toolchains()

load("@googleapis//:repository_rules.bzl", "switched_rules_by_language")

switched_rules_by_language(
    name = "com_google_googleapis_imports",
)

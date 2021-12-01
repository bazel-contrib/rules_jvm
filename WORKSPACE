workspace(name = "rules_jvm_contrib")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

RULES_JVM_EXTERNAL_TAG = "4.2"

RULES_JVM_EXTERNAL_SHA = "cd1a77b7b02e8e008439ca76fd34f5b07aecb8c752961f9640dea15e9e5ba1ca"

http_archive(
    name = "rules_jvm_external",
    sha256 = RULES_JVM_EXTERNAL_SHA,
    strip_prefix = "rules_jvm_external-%s" % RULES_JVM_EXTERNAL_TAG,
    url = "https://github.com/bazelbuild/rules_jvm_external/archive/%s.zip" % RULES_JVM_EXTERNAL_TAG,
)

load("@rules_jvm_external//:defs.bzl", "maven_install")

maven_install(
    name = "rules_jvm_contrib_deps",
    artifacts = [
        # These can be versioned independently of the versions in `repositories.bzl`
        # so long as the version numbers are higher.
        "org.junit.jupiter:junit-jupiter-engine:5.8.1",
        "org.junit.jupiter:junit-jupiter-api:5.8.1",
        "org.junit.platform:junit-platform-launcher:1.8.1",
        "org.junit.platform:junit-platform-reporting:1.8.1",
        "org.junit.vintage:junit-vintage-engine:5.8.1",
    ],
    fetch_sources = True,
    repositories = [
        "https://repo1.maven.org/maven2",
    ],
)

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
    fetch_sources = True,
    repositories = [
        "https://repo1.maven.org/maven2",
    ],
)

workspace(name = "rules_jvm_contrib")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

load("//:repositories.bzl", "rules_jvm_contrib_deps")

rules_jvm_contrib_deps()

load("//:setup.bzl", "rules_jvm_contrib_setup")

rules_jvm_contrib_setup()

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
    ],
    fetch_sources = True,
    fail_if_repin_required = True,
    maven_install_json = "@rules_jvm_contrib//:frozen_deps_install.json",
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
    fetch_sources = True,
    fail_if_repin_required = True,
    maven_install_json = "@//:maven_install.json",
    repositories = [
        "https://repo1.maven.org/maven2",
    ],
)

load("@maven//:defs.bzl", "pinned_maven_install")
pinned_maven_install()


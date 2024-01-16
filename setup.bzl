load("@apple_rules_lint//lint:setup.bzl", "ruleset_lint_setup")
load("@contrib_rules_jvm_deps//:compat.bzl", "compat_repositories")
load("@contrib_rules_jvm_deps//:defs.bzl", "pinned_maven_install")
load("@rules_java//java:repositories.bzl", "rules_java_dependencies", "rules_java_toolchains")

def contrib_rules_jvm_setup():
    ruleset_lint_setup()

    # When using bazel 5, we have undefined toolchains from rules_java. This should be fine to skip, since we only need
    # it for the `JavaInfo` definition.
    major_version = native.bazel_version.partition(".")[0]
    if major_version != "5":
        rules_java_dependencies()
        rules_java_toolchains()

    pinned_maven_install()
    compat_repositories()

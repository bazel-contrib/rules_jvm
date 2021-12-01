load("@apple_rules_lint//lint:setup.bzl", "ruleset_lint_setup")
load("@rules_jvm_contrib_deps//:defs.bzl", "pinned_maven_install")

def rules_jvm_contrib_setup():
    ruleset_lint_setup()
    pinned_maven_install()

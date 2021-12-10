load("@apple_rules_lint//lint:setup.bzl", "ruleset_lint_setup")
load("@contrib_rules_jvm_deps//:defs.bzl", "pinned_maven_install")

def contrib_rules_jvm_setup():
    ruleset_lint_setup()
    pinned_maven_install()

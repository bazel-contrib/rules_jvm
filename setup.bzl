load("@rules_jvm_contrib_deps//:defs.bzl", "pinned_maven_install")

def rules_jvm_contrib_setup():
    pinned_maven_install()

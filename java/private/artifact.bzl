load("@rules_jvm_external//:defs.bzl", _artifact = "artifact")

def artifact(coords):
    return _artifact(coords, repository_name = "contrib_rules_jvm_deps")

load("@apple_rules_lint//lint:setup.bzl", "ruleset_lint_setup")
load("@contrib_rules_jvm_deps//:compat.bzl", "compat_repositories")
load("@contrib_rules_jvm_deps//:defs.bzl", "pinned_maven_install")
load("@io_grpc_grpc_java//:repositories.bzl", "io_grpc_grpc_proto")
load("//third_party/go:repositories.bzl", "go_deps")

def contrib_rules_jvm_setup():
    ruleset_lint_setup()
    pinned_maven_install()
    compat_repositories()
    io_grpc_grpc_proto()
    go_deps()

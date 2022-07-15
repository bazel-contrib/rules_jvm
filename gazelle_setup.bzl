load("@io_grpc_grpc_java//:repositories.bzl", "io_grpc_grpc_proto")
load("//third_party/go:repositories.bzl", "go_deps")

def contrib_rules_jvm_gazelle_setup():
    io_grpc_grpc_proto()
    go_deps()

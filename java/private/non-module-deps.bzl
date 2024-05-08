load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("//:repositories.bzl", "io_grpc_grpc_java")

def _non_module_dependencies_impl(mctx):
    io_grpc_grpc_java()

non_module_deps = module_extension(
    implementation = _non_module_dependencies_impl,
)

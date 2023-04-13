load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

def _non_module_dependencies_impl(mctx):
    maybe(
        http_archive,
        name = "io_grpc_grpc_java",
        sha256 = "b6cfc524647cc680e66989ab22a10b66dc5de8c6d8499f91a7e633634c594c61",
        strip_prefix = "grpc-java-1.51.1",
        urls = ["https://github.com/grpc/grpc-java/archive/v1.51.1.tar.gz"],
    )

non_module_deps = module_extension(
    implementation = _non_module_dependencies_impl,
)

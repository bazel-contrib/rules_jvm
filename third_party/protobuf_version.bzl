# Please keep in sync with MODULE.bazel until
# https://github.com/bazelbuild/bazel/issues/17880 is solved.
PROTOBUF_VERSION = "21.7"

# The java packages are published to maven under a different versioning scheme.
PROTOBUF_JAVA_VERSION = "3.{}".format(PROTOBUF_VERSION)

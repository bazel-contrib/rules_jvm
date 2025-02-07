#!/usr/bin/env bash
set -eufo pipefail

# Update the Java bits
REPIN=1 bazel run @contrib_rules_jvm_deps//:pin
REPIN=1 bazel run @contrib_rules_jvm_tests//:pin
bazel run //tools:freeze-deps

# And now the Go bits
bazel run @rules_go//go -- mod tidy
bazel run //:buildifier

./tools/format.sh

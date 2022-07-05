#!/usr/bin/env bash
set -eufo pipefail

go mod tidy

# Work around https://github.com/bazelbuild/bazel-gazelle/issues/999
# Ideally we would delete the below block and replace it with this commented command:
#bazel run //:gazelle_go -- update-repos \
#    -from_file=go.mod \
#    -prune \
#    -to_macro "third_party/go/repositories.bzl%go_deps"
GO_DEPS_FILE="third_party/go/repositories.bzl"
bazel run //:gazelle -- update-repos -from_file=go.mod -prune -to_macro "${GO_DEPS_FILE}%go_deps"
sed '/^$/d' "$GO_DEPS_FILE" >"${GO_DEPS_FILE}.new"
mv "${GO_DEPS_FILE}.new" "$GO_DEPS_FILE"
bazel run //:buildifier


REPIN=1 bazel run @unpinned_maven//:pin
./tools/freeze-deps.py
bazel fetch "@frozen_deps//..."

./tools/format.sh

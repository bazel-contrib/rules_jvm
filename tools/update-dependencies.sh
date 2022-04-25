#!/usr/bin/env bash
set -eufo pipefail

go mod tidy
bazel run //:gazelle_go -- update-repos \
    -from_file=go.mod \
    -prune \
    -to_macro "third_party/go/repositories.bzl%go_deps"

REPIN=1 bazel run @unpinned_maven//:pin
./tools/freeze-deps.py
bazel fetch "@frozen_deps//..."
bazel run //java/gazelle/cmd/parsejars -- \
    --maven-install "frozen_deps_install.json" \
    --output "frozen_deps_manifest.json" \
    --repo-root "$PWD"

./tools/format.sh

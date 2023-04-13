#!/usr/bin/env bash
set -eufo pipefail

bazel build --noenable_bzlmod //docs:readme
cp -f bazel-bin/docs/README.md README.md

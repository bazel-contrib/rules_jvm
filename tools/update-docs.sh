#!/usr/bin/env bash
set -eufo pipefail

bazel build //docs:readme
cp -f bazel-bin/docs/README.md README.md

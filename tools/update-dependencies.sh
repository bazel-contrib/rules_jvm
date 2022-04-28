#!/usr/bin/env bash
set -eufo pipefail

./tools/freeze-deps.py
REPIN=1 bazel run @unpinned_maven//:pin

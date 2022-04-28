#!/usr/bin/env bash
set -eufo pipefail

GOOGLE_JAVA_FORMAT="$(bazel run --run_under=echo //tools:google-java-format)"

find "$PWD" -type f -name '*.java' | xargs "$GOOGLE_JAVA_FORMAT" --replace
bazel run //:gazelle

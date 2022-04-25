#!/usr/bin/env bash
# Code formatter.
set -eufo pipefail

section() {
    echo "- $*" >&2
}

GOIMPORTS="$(bazel run --run_under=echo @org_golang_x_tools//cmd/goimports)"
GOOGLE_JAVA_FORMAT="$(bazel run --run_under=echo //tools:google-java-format)"
IMPORTSORT="$(bazel run --run_under=echo @com_github_aristanetworks_goarista//cmd/importsort)"

section "Go"
echo "    goimports" >&2
find "$PWD" -type f -name '*.go' | grep -v ".pb.go" | xargs "$GOIMPORTS" -l -w
echo "    importsort" >&2
find "$PWD" -type f -name '*.go' | grep -v ".pb.go" | xargs "$IMPORTSORT" -w -s NOT_SPECIFIED

section "Java"
echo "    google-java-format" >&2
find "$PWD" -type f -name '*.java' | grep -v "testdata\|generators\/workspace" | xargs "$GOOGLE_JAVA_FORMAT" --replace

section "Build files"
echo "    gazelle" >&2
bazel run //:gazelle
echo "    buildifier" >&2
bazel run //:buildifier

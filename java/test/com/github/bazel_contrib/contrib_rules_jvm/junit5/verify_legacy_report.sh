#!/bin/bash
# Verifies that running legacy-report-generator-test with report_generator="legacy"
# produces a correctly-structured JUnit legacy XML report in TEST_UNDECLARED_OUTPUTS_DIR.
#
# $1 is the rlocationpath of the legacy-report-generator-test binary, passed via args
# in the sh_test BUILD target.
# $2 is the rlocationpath of the legacy-report-verifier binary.

set -euo pipefail

readonly BINARY="$TEST_SRCDIR/$1"
readonly VERIFIER="$TEST_SRCDIR/$2"

# Use a separate XML output file so the nested binary doesn't overwrite the
# sh_test's own Bazel XML test report.
XML_OUTPUT="$(mktemp --suffix=.xml)"

# Run the binary in the Bazel test environment, sharing TEST_UNDECLARED_OUTPUTS_DIR
# so that the legacy report it generates is written to the directory Bazel manages.
# Unset TEST_PREMATURE_EXIT_FILE so the nested binary doesn't interact with the
# outer test's premature-exit sentinel.
(
  unset TEST_PREMATURE_EXIT_FILE || true
  export XML_OUTPUT_FILE="$XML_OUTPUT"
  "$BINARY"
)

# The legacy reporter writes one TEST-<classname>.xml file per test class.
REPORT="$(find "$TEST_UNDECLARED_OUTPUTS_DIR" -name "TEST-*.xml" | head -1)"
if [[ -z "$REPORT" ]]; then
  echo "FAIL: No legacy XML report file (TEST-*.xml) found in $TEST_UNDECLARED_OUTPUTS_DIR"
  ls -la "$TEST_UNDECLARED_OUTPUTS_DIR" || true
  exit 1
fi
echo "Found report: $REPORT"

"$VERIFIER" "$REPORT"

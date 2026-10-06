#!/bin/bash
# Verifies that running default-report-generator-test without report_generator:
#   - writes no files to TEST_UNDECLARED_OUTPUTS_DIR
#   - writes a correctly-structured JUnit XML report to XML_OUTPUT_FILE
#
# $1 is the rlocationpath of the default-report-generator-test binary, passed via args
# in the sh_test BUILD target.
# $2 is the rlocationpath of the default-report-verifier binary.

set -euo pipefail

readonly BINARY="$TEST_SRCDIR/$1"
readonly VERIFIER="$TEST_SRCDIR/$2"

# Route the inner binary's XML output to a temp file rather than the outer test's report.
XML_OUTPUT="$(mktemp --suffix=.xml)"

(
  unset TEST_PREMATURE_EXIT_FILE || true
  export XML_OUTPUT_FILE="$XML_OUTPUT"
  "$BINARY"
)

# No additional report files should have been written to TEST_UNDECLARED_OUTPUTS_DIR.
if [[ -n "$(find "$TEST_UNDECLARED_OUTPUTS_DIR" -mindepth 1 -print -quit)" ]]; then
  echo "FAIL: Files were unexpectedly written to TEST_UNDECLARED_OUTPUTS_DIR"
  ls -la "$TEST_UNDECLARED_OUTPUTS_DIR"
  exit 1
fi
echo "PASS: TEST_UNDECLARED_OUTPUTS_DIR is empty (no additional report generated)"

# Verify the standard XML output written to XML_OUTPUT_FILE.
"$VERIFIER" "$XML_OUTPUT"

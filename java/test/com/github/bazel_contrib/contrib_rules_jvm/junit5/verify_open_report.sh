#!/bin/bash
# Verifies that running open-report-generator-test with report_generator="open"
# produces an Open Test Reporting XML file in TEST_UNDECLARED_OUTPUTS_DIR.
#
# $1 is the rlocationpath of the open-report-generator-test binary, passed via args
# in the sh_test BUILD target.
# $2 is the rlocationpath of the open-report-verifier binary.

set -euo pipefail

readonly BINARY="$TEST_SRCDIR/$1"
readonly VERIFIER="$TEST_SRCDIR/$2"

XML_OUTPUT="$(mktemp --suffix=.xml)"

(
  unset TEST_PREMATURE_EXIT_FILE || true
  export XML_OUTPUT_FILE="$XML_OUTPUT"
  "$BINARY"
)

# The Open Test Reporting format writes junit-platform-events-*.xml files into the
# output directory.
REPORT="$(find "$TEST_UNDECLARED_OUTPUTS_DIR" -name "junit-platform-events-*.xml" | head -1)"
if [[ -z "$REPORT" ]]; then
  echo "FAIL: No junit-platform-events-*.xml file found in $TEST_UNDECLARED_OUTPUTS_DIR"
  ls -la "$TEST_UNDECLARED_OUTPUTS_DIR" || true
  exit 1
fi
echo "Found report: $REPORT"

if [[ ! -s "$REPORT" ]]; then
  echo "FAIL: Report file is empty: $REPORT"
  exit 1
fi

"$VERIFIER" "$REPORT"

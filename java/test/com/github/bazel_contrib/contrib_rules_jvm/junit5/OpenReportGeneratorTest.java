package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;

public class OpenReportGeneratorTest {

  @Test
  public void shouldUseOpenReportGenerator() {
    // This test verifies that ActualRunner doesn't crash when report_generator = "open".
    // The report file written to TEST_UNDECLARED_OUTPUTS_DIR is verified by the
    // verify-open-report sh_test target, which runs this binary and inspects its output.
    assertTrue(true, "Test executed successfully with open report generator");
  }
}

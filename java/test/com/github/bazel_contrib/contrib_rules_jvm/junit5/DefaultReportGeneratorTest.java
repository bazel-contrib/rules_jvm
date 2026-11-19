package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;

public class DefaultReportGeneratorTest {

  @Test
  public void shouldUseDefaultReportGenerator() {
    // This test verifies the default behavior when report_generator is not specified.
    // The standard Bazel XML output written to XML_OUTPUT_FILE is verified by the
    // verify-default-report sh_test target, which also checks that no additional
    // files are written to TEST_UNDECLARED_OUTPUTS_DIR.
    assertTrue(true, "Test executed successfully with default report generator");
  }
}

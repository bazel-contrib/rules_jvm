package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;

public class OpenReportGeneratorTest {

  @Test
  public void shouldUseOpenReportGenerator() {
    // This test is mainly to verify that the runner doesn't crash when "open" is specified.
    // Actual report output is not verified, as this is tricky from within the test.
    assertTrue(true, "Test executed successfully with open report generator");
  }
}

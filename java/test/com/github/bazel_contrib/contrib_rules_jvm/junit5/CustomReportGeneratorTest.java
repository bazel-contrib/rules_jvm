package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import org.junit.jupiter.api.Test;

public class CustomReportGeneratorTest {

  @Test
  public void shouldUseCustomReportGenerator() throws IOException {
    // CustomReportGenerator creates this marker in its constructor, which runs before any tests.
    Path markerFile = Path.of("custom_report_generator_marker.txt");
    assertTrue(
        Files.exists(markerFile),
        "Marker file should exist, indicating CustomReportGenerator was instantiated");

    // CustomReportGenerator writes its report in testPlanExecutionStarted, which also fires
    // before any test methods, so the file is already present here.
    String outputDirStr = System.getenv("TEST_UNDECLARED_OUTPUTS_DIR");
    assertNotNull(outputDirStr, "TEST_UNDECLARED_OUTPUTS_DIR should be set by the Bazel test runner");
    Path reportFile = Path.of(outputDirStr).resolve("custom-report.txt");
    assertTrue(Files.exists(reportFile), "Custom report file should exist at: " + reportFile);
    assertEquals(
        "Executing Test Plan",
        Files.readString(reportFile),
        "Custom report should contain the expected content");
  }
}

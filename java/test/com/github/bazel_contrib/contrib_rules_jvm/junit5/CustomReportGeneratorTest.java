package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import org.junit.jupiter.api.Test;

public class CustomReportGeneratorTest {

  @Test
  public void shouldUseCustomReportGenerator() throws IOException {
    // This test verifies that the custom report generator was used.
    // The custom generator writes a specific marker file when it starts.
    Path markerFile = Path.of("custom_report_generator_marker.txt");
    assertTrue(Files.exists(markerFile), "Marker file should exist, indicating CustomReportGenerator was used");
  }
}

package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.platform.launcher.TestExecutionListener;
import org.junit.platform.launcher.TestPlan;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

public class CustomReportGenerator implements TestExecutionListener {
  private Path outFile;

  public CustomReportGenerator(Path outputDir) {
    outFile = outputDir.resolve("custom-report.txt");

    try {
      Files.createFile(Path.of("custom_report_generator_marker.txt"));
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  @Override
  public void testPlanExecutionStarted(TestPlan testPlan) {
    System.out.println("USING CUSTOM REPORT");
    try {
      Files.writeString(outFile, "Executing Test Plan");
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }
}

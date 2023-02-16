package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static java.nio.file.StandardOpenOption.DELETE_ON_CLOSE;
import static java.nio.file.StandardOpenOption.TRUNCATE_EXISTING;
import static java.nio.file.StandardOpenOption.WRITE;
import static org.junit.platform.launcher.EngineFilter.includeEngines;
import static org.junit.platform.launcher.EngineFilter.excludeEngines;

import java.io.File;
import java.io.IOException;
import java.io.PrintWriter;
import java.io.UncheckedIOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;
import org.junit.platform.engine.discovery.DiscoverySelectors;
import org.junit.platform.launcher.LauncherConstants;
import org.junit.platform.launcher.core.LauncherConfig;
import org.junit.platform.launcher.core.LauncherDiscoveryRequestBuilder;
import org.junit.platform.launcher.core.LauncherFactory;

public class ActualRunner implements RunsTest {

  @Override
  public boolean run(String testClassName) {
    var out = System.getenv("XML_OUTPUT_FILE");
    Path xmlOut;
    try {
      xmlOut = out != null ? Paths.get(out) : Files.createTempFile("test", ".xml");
      Files.createDirectories(xmlOut.getParent());
    } catch (IOException e) {
      throw new UncheckedIOException(e);
    }

    try (var bazelJunitXml = new BazelJUnitOutputListener(xmlOut)) {
      var summary = new CommandLineSummary();

      LauncherConfig config =
          LauncherConfig.builder().addTestExecutionListeners(bazelJunitXml, summary).build();

      var classSelector = DiscoverySelectors.selectClass(testClassName);

      var request =
          LauncherDiscoveryRequestBuilder.request()
              .selectors(List.of(classSelector))
              .configurationParameter(LauncherConstants.CAPTURE_STDERR_PROPERTY_NAME, "true")
              .configurationParameter(LauncherConstants.CAPTURE_STDOUT_PROPERTY_NAME, "true");

      String filter = System.getenv("TESTBRIDGE_TEST_ONLY");
      request.filters(new PatternFilter(filter));

      List<String> includeEngines = System.getProperty("bazel.include_engines") == null ? null : List.of(System.getProperty("bazel.include_engines").split(","));
      List<String> excludeEngines = System.getProperty("bazel.exclude_engines") == null ? null : List.of(System.getProperty("bazel.exclude_engines").split(","));
      if (includeEngines != null) {
        request.filters(includeEngines(includeEngines));
      }
      if (excludeEngines != null) {
        request.filters(excludeEngines(excludeEngines));
      }

      File exitFile = getExitFile();

      var launcher = LauncherFactory.create(config);
      launcher.execute(request.build());

      deleteExitFile(exitFile);

      try (PrintWriter writer = new PrintWriter(System.out)) {
        summary.writeTo(writer);
      }

      return summary.getFailureCount() == 0;
    }
  }

  private File getExitFile() {
    String exitFileName = System.getenv("TEST_PREMATURE_EXIT_FILE");
    if (exitFileName == null) {
      return null;
    }

    File exitFile = new File(exitFileName);
    try {
      Files.write(exitFile.toPath(), "".getBytes(), WRITE, DELETE_ON_CLOSE, TRUNCATE_EXISTING);
    } catch (IOException e) {
      return null;
    }

    return exitFile;
  }

  private void deleteExitFile(File exitFile) {
    if (exitFile != null) {
      try {
        exitFile.delete();
      } catch (Throwable t) {
        // Ignore.
      }
    }
  }
}

package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static java.nio.file.StandardOpenOption.DELETE_ON_CLOSE;
import static java.nio.file.StandardOpenOption.TRUNCATE_EXISTING;
import static java.nio.file.StandardOpenOption.WRITE;
import static org.junit.platform.launcher.EngineFilter.excludeEngines;
import static org.junit.platform.launcher.EngineFilter.includeEngines;

import java.io.File;
import java.io.IOException;
import java.io.PrintWriter;
import java.io.UncheckedIOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import org.junit.platform.engine.DiscoverySelector;
import org.junit.platform.engine.discovery.DiscoverySelectors;
import org.junit.platform.launcher.Launcher;
import org.junit.platform.launcher.LauncherConstants;
import org.junit.platform.launcher.TagFilter;
import org.junit.platform.launcher.core.LauncherConfig;
import org.junit.platform.launcher.core.LauncherDiscoveryRequestBuilder;
import org.junit.platform.launcher.core.LauncherFactory;

public class ActualRunner implements RunsTest {

  @Override
  public boolean run(String testClassName) {
    String out = System.getenv("XML_OUTPUT_FILE");
    Path xmlOut;
    try {
      xmlOut = out != null ? Paths.get(out) : Files.createTempFile("test", ".xml");
      Files.createDirectories(xmlOut.getParent());
    } catch (IOException e) {
      throw new UncheckedIOException(e);
    }

    try (BazelJUnitOutputListener bazelJunitXml = new BazelJUnitOutputListener(xmlOut)) {
      CommandLineSummary summary = new CommandLineSummary();

      LauncherConfig config =
          LauncherConfig.builder().addTestExecutionListeners(bazelJunitXml, summary).build();

      DiscoverySelector classSelector = DiscoverySelectors.selectClass(testClassName);

      LauncherDiscoveryRequestBuilder request =
          LauncherDiscoveryRequestBuilder.request()
              .selectors(Collections.singletonList(classSelector))
              .configurationParameter(LauncherConstants.CAPTURE_STDERR_PROPERTY_NAME, "true")
              .configurationParameter(LauncherConstants.CAPTURE_STDOUT_PROPERTY_NAME, "true");

      String filter = System.getenv("TESTBRIDGE_TEST_ONLY");
      request.filters(new PatternFilter(filter));

      String includeTags = System.getProperty("JUNIT5_INCLUDE_TAGS");
      if (includeTags != null && !includeTags.isEmpty()) {
        request.filters(TagFilter.includeTags(includeTags.split(",")));
      }

      String excludeTags = System.getProperty("JUNIT5_EXCLUDE_TAGS");
      if (excludeTags != null && !excludeTags.isEmpty()) {
        request.filters(TagFilter.excludeTags(excludeTags.split(",")));
      }

      List<String> includeEngines =
          System.getProperty("JUNIT5_INCLUDE_ENGINES") == null
              ? null
              : Arrays.asList(System.getProperty("JUNIT5_INCLUDE_ENGINES").split(","));
      List<String> excludeEngines =
          System.getProperty("JUNIT5_EXCLUDE_ENGINES") == null
              ? null
              : Arrays.asList(System.getProperty("JUNIT5_EXCLUDE_ENGINES").split(","));
      if (includeEngines != null) {
        request.filters(includeEngines(includeEngines));
      }
      if (excludeEngines != null) {
        request.filters(excludeEngines(excludeEngines));
      }

      File exitFile = getExitFile();

      Launcher launcher = LauncherFactory.create(config);
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

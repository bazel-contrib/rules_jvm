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
import java.lang.annotation.Annotation;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Modifier;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;
import org.junit.jupiter.engine.Constants;
import org.junit.platform.engine.DiscoverySelector;
import org.junit.platform.engine.discovery.DiscoverySelectors;
import org.junit.platform.launcher.Launcher;
import org.junit.platform.launcher.LauncherConstants;
import org.junit.platform.launcher.TagFilter;
import org.junit.platform.launcher.TestExecutionListener;
import org.junit.platform.launcher.core.LauncherConfig;
import org.junit.platform.launcher.core.LauncherDiscoveryRequestBuilder;
import org.junit.platform.launcher.core.LauncherFactory;

public class ActualRunner implements RunsTest {

  public static final String REPORT_TYPE_LEGACY = "legacy";
  public static final String REPORT_TYPE_OPEN = "open";

  @Override
  public boolean run(String testClassName) {
    Path xmlOut = getTestXmlFile();

    try (BazelJUnitOutputListener bazelJUnitXml = new BazelJUnitOutputListener(xmlOut)) {
      Runtime.getRuntime()
          .addShutdownHook(
              new Thread(
                  () -> {
                    bazelJUnitXml.closeForInterrupt();
                  }));

      CommandLineSummary summary = new CommandLineSummary();
      FailFastExtension failFastExtension = new FailFastExtension();

      LauncherConfig.Builder configBuilder = LauncherConfig.builder()
              .addTestExecutionListeners(bazelJUnitXml, summary, failFastExtension)
              .addPostDiscoveryFilters(TestSharding.makeShardFilter());

      final Class<?> testClass;
      try {
        testClass = Class.forName(testClassName, false, getClass().getClassLoader());
      } catch (ClassNotFoundException e) {
        throw new RuntimeException("Failed to find testClass", e);
      }

      // We only allow for one level of nesting at the moment
      boolean enclosed = isRunWithEnclosed(testClass);
      List<DiscoverySelector> classSelectors =
          enclosed
              ? new ArrayList<>()
              : Arrays.stream(testClass.getDeclaredClasses())
                  .filter(clazz -> Modifier.isStatic(clazz.getModifiers()))
                  .map(DiscoverySelectors::selectClass)
                  .collect(Collectors.toList());

      classSelectors.add(DiscoverySelectors.selectClass(testClassName));

      LauncherDiscoveryRequestBuilder request =
          LauncherDiscoveryRequestBuilder.request()
              .selectors(classSelectors)
              .configurationParameter(LauncherConstants.CAPTURE_STDERR_PROPERTY_NAME, "true")
              .configurationParameter(LauncherConstants.CAPTURE_STDOUT_PROPERTY_NAME, "true")
              .configurationParameter(
                  Constants.EXTENSIONS_AUTODETECTION_ENABLED_PROPERTY_NAME, "true");

      configureReportGenerators(request, configBuilder);

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

      Launcher launcher = LauncherFactory.create(configBuilder.build());
      launcher.execute(request.build());

      deleteExitFile(exitFile);

      try (PrintWriter writer = new PrintWriter(System.out)) {
        summary.writeTo(writer);
      }

      return summary.getFailureCount() == 0;
    }
  }

  private static void configureReportGenerators(LauncherDiscoveryRequestBuilder request, LauncherConfig.Builder configBuilder) {
    String reportGenerator = System.getProperty("JUNIT5_REPORT_GENERATOR");

    if (reportGenerator != null) {
      Path undeclaredOutDir = getUndeclaredOutputDirectory();

      switch (reportGenerator) {
        case REPORT_TYPE_LEGACY:
          try {
            Class<?> clazz = Class.forName("org.junit.platform.reporting.legacy.xml.LegacyXmlReportGeneratingListener");
            configBuilder.addTestExecutionListeners((TestExecutionListener) clazz.getConstructor(Path.class, PrintWriter.class)
                    .newInstance(undeclaredOutDir, new PrintWriter(System.out)));
            request.configurationParameter("junit.platform.reporting.output.dir", undeclaredOutDir.toString());
          } catch (ReflectiveOperationException e) {
            throw new RuntimeException("Legacy Test Reporting is only available in JUnit >= 5.4" + reportGenerator, e);
          }
          break;
        case REPORT_TYPE_OPEN:
          try {
            Class<?> clazz = Class.forName("org.junit.platform.reporting.open.xml.OpenTestReportGeneratingListener");
            configBuilder.addTestExecutionListeners((TestExecutionListener) clazz.getConstructor().newInstance());
            request.configurationParameter("junit.platform.reporting.open.xml.enabled", "true");
            request.configurationParameter("junit.platform.reporting.output.dir", undeclaredOutDir.toString());
          } catch (ReflectiveOperationException e) {
            throw new RuntimeException("Open Test Reporting is only available in JUnit >= 5.9" + reportGenerator, e);
          }
          break;
        default:
          // We assume report_generator is a classname
          try {
            Class<?> clazz = Class.forName(reportGenerator);
            configBuilder.addTestExecutionListeners((TestExecutionListener) clazz.getConstructor(Path.class).newInstance(undeclaredOutDir));
          } catch (ReflectiveOperationException e) {
            throw new RuntimeException("Failed to create report generator class: " + reportGenerator, e);
          }
      }
    }
  }

  /**
   * Checks if the test class is annotation with `@RunWith(Enclosed.class)`. We deliberately avoid
   * using types here to avoid polluting the classpath with junit4 deps.
   */
  private boolean isRunWithEnclosed(Class<?> clazz) {
    for (Annotation annotation : clazz.getAnnotations()) {
      Class<? extends Annotation> type = annotation.annotationType();
      if (type.getName().equals("org.junit.runner.RunWith")) {
        try {
          Class<?> runner = (Class<?>) type.getMethod("value").invoke(annotation, (Object[]) null);
          if (runner.getName().equals("org.junit.experimental.runners.Enclosed")) {
            return true;
          }
        } catch (NoSuchMethodException | IllegalAccessException | InvocationTargetException e) {
          return false;
        }
      }
    }
    return false;
  }

  private static Path getUndeclaredOutputDirectory() {
    String dirStr = System.getenv("TEST_UNDECLARED_OUTPUTS_DIR");
    Path dirPath;
    try {
      dirPath = dirStr != null ? Paths.get(dirStr) : Files.createTempDirectory("test.outputs");
      Files.createDirectories(dirPath);
    } catch (IOException e) {
      throw new UncheckedIOException(e);
    }
    return dirPath;
  }

  private static Path getTestXmlFile() {
    String fileStr = System.getenv("XML_OUTPUT_FILE");
    Path filePath;
    try {
      filePath = fileStr != null ? Paths.get(fileStr) : Files.createTempFile("test", ".xml");
      Files.createDirectories(filePath.getParent());
    } catch (IOException e) {
      throw new UncheckedIOException(e);
    }
    return filePath;
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

package com.github.bazel_contrib.contrib_rules_jvm.junit5;

/**
 * Test bootstrapper. This class only depends on the JRE (java 11+) and will ensure that the
 * required dependencies for a junit5 test are on the classpath before creating the actual runner.
 * In this way we can offer a useful error message to people, which is always nice, right?
 *
 * <p>Most of the configuration information can be found on this page in the <a
 * href="https://docs.bazel.build/versions/master/test-encyclopedia.html">Test Encyclopedia</a>.
 */
public class JUnit5Runner {

  private static final String JUNIT5_RUNNER_CLASS =
      "com.github.bazel_contrib.contrib_rules_jvm.junit5.ActualRunner";

  private static final String JAVA17_SYSTEM_EXIT_TOGGLE =
      "com.github.bazel_contrib.contrib_rules_jvm.junit5.Java17SystemExitToggle";

  public static void main(String[] args) {
    var testSuite = System.getProperty("bazel.test_suite");

    var systemExitToggle = getSystemExitToggle();

    if (testSuite == null || testSuite.isBlank()) {
      System.err.println("No test suite specified");
      exit(systemExitToggle, 2); // Same error code as Bazel's own test runner
    }

    detectJUnit5Classes();

    systemExitToggle.prevent();

    try {
      var constructor =
          Class.forName(JUNIT5_RUNNER_CLASS).asSubclass(RunsTest.class).getConstructor();
      var runsTest = constructor.newInstance();
      if (!runsTest.run(testSuite)) {
        exit(systemExitToggle, 2);
      }
    } catch (ReflectiveOperationException e) {
      e.printStackTrace(System.err);
      System.err.println("Unable to create delegate test runner");
      exit(systemExitToggle, 2);
    }

    // Exit manually. If we don't do this then tests which hold resources
    // such as Threads may prevent us from exiting properly.
    exit(systemExitToggle, 0);
  }

  private static SystemExitToggle getSystemExitToggle() {
    var version = Runtime.version();

    if (version.feature() < 12) {
      return new Java11SystemExitToggle();
    }

    // Load the java 17 toggle by reflection, because it's tied
    // so closely to the OpenJDK (it relies on the internal fields
    // of both `sun.misc.Unsafe` and `java.lang.System`: it's a
    // gross hack.
    try {
      var java17ToggleClazz =
          Class.forName(JAVA17_SYSTEM_EXIT_TOGGLE).asSubclass(SystemExitToggle.class);
      return java17ToggleClazz.getDeclaredConstructor().newInstance();
    } catch (Exception e) {
      // We don't care _why_ we can't load the toggle, but we can't. Ideally
      // this would be a combination of `ReflectiveOperationException` and
      // `SecurityException`, but the latter is scheduled for removal so
      // relying on it seems like a Bad Idea.
      System.err.println("Failed to load Java 17 system exit override: " + e.getMessage());

      // Fall through
    }

    System.err.println(
        "Unable to create a mechanism to prevent `System.exit` being called. "
            + "Tests may cause `bazel test` to exit prematurely.");

    return new NullSystemExitToggle();
  }

  private static void detectJUnit5Classes() {
    checkClass(
        "org.junit.jupiter.engine.JupiterTestEngine", "org.junit.jupiter:junit-jupiter-engine");
    checkClass(
        "org.junit.platform.commons.JUnitException", "org.junit.platform:junit-platform-commons");
    checkClass(
        "org.junit.platform.engine.ExecutionRequest", "org.junit.platform:junit-platform-engine");
    checkClass(
        "org.junit.platform.launcher.TestPlan", "org.junit.platform:junit-platform-launcher");
    checkClass(
        "org.junit.platform.reporting.legacy.LegacyReportingUtils",
        "org.junit.platform:junit-platform-reporting");
  }

  private static void checkClass(String className, String containedInDependency) {
    try {
      Class.forName(className);
    } catch (ReflectiveOperationException e) {
      throw new IllegalStateException(
          String.format(
              "JUnit 5 test runner is missing a dependency on `artifact(\"%s\")`%n",
              containedInDependency));
    }
  }

  private static void exit(SystemExitToggle toggle, int value) {
    toggle.allow();
    System.exit(value);
  }
}

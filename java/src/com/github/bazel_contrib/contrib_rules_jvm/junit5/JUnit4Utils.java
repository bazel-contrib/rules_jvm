package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.opentest4j.TestAbortedException;

/** Used to prevent a hard dependency on JUnit4 when using the JUnit5 runner. */
public class JUnit4Utils {

  private static final Class<?> JUNIT4_ASSUMPTION_CLASS;

  static {
    Class<?> clazz = null;
    try {
      clazz = Class.forName("org.junit.AssumptionViolatedException");
    } catch (ClassNotFoundException e) {
      // This is fine. It just means that JUnit 4 isn't on the classpath
    }
    JUNIT4_ASSUMPTION_CLASS = clazz;
  }

  private JUnit4Utils() {
    // Utility class
  }

  public static boolean isReasonToSkipTest(Throwable throwable) {
    if (throwable instanceof TestAbortedException) {
      return true;
    }

    if (JUNIT4_ASSUMPTION_CLASS == null) {
      return false;
    }

    return JUNIT4_ASSUMPTION_CLASS.isAssignableFrom(throwable.getClass());
  }
}

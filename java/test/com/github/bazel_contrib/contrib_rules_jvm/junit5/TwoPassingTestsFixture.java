package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.Test;

/**
 * A fixture class with two passing {@code @Test} methods, used by {@link CommandLineSummaryTest}.
 * Deliberately does not end in "Test.java" so it is not itself picked up as a test by {@code
 * java_test_suite}.
 */
public class TwoPassingTestsFixture {
  @Test
  public void first() {}

  @Test
  public void second() {}
}

package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.Test;

/**
 * A fixture class whose test fails with an {@link OutOfMemoryError} wrapped in another throwable,
 * used by {@link CommandLineSummaryTest}. Deliberately does not end in "Test.java" so it is not
 * itself picked up as a test by {@code java_test_suite}.
 */
public class WrappedOomFailureFixture {
  @Test
  public void shouldFailWithWrappedOom() {
    throw new IllegalStateException("wrapped", new OutOfMemoryError("wrapped OOM"));
  }
}

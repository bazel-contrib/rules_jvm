package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.Test;

/**
 * A fixture class whose test fails with an ordinary exception in a cyclic cause chain, used by
 * {@link CommandLineSummaryTest}. Deliberately does not end in "Test.java" so it is not itself
 * picked up as a test by {@code java_test_suite}.
 */
public class CyclicNonOomFailureFixture {
  @Test
  public void shouldFailWithCyclicCause() {
    IllegalStateException failure = new IllegalStateException("cyclic normal failure");
    IllegalArgumentException cause = new IllegalArgumentException("cyclic normal cause");
    failure.initCause(cause);
    cause.initCause(failure);
    throw failure;
  }
}

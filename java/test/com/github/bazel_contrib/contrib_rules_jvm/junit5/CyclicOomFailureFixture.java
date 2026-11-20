package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.Test;

/**
 * A fixture class whose test fails with an {@link OutOfMemoryError} in a cyclic cause chain, used
 * by {@link CommandLineSummaryTest}. Deliberately does not end in "Test.java" so it is not itself
 * picked up as a test by {@code java_test_suite}.
 */
public class CyclicOomFailureFixture {
  @Test
  public void shouldFailWithCyclicOom() {
    IllegalStateException failure = new IllegalStateException("cyclic failure");
    OutOfMemoryError oom = new OutOfMemoryError("cyclic OOM");
    failure.initCause(oom);
    oom.initCause(failure);
    throw failure;
  }
}

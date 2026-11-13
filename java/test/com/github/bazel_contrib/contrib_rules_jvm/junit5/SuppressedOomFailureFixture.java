package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.Test;

/**
 * A fixture class whose test fails with an ordinary exception while an {@link OutOfMemoryError}
 * thrown by a resource's {@code close()} is attached to it as a suppressed exception, used by
 * {@link CommandLineSummaryTest}. Deliberately does not end in "Test.java" so it is not itself
 * picked up as a test by {@code java_test_suite}.
 */
public class SuppressedOomFailureFixture {
  @Test
  public void shouldFailWithOomSuppressedByTryWithResources() throws Exception {
    try (AutoCloseable resource =
        () -> {
          throw new OutOfMemoryError("suppressed OOM");
        }) {
      throw new IllegalStateException("failure before cleanup OOM");
    }
  }
}

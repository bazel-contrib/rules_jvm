package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertNotEquals;

import org.junit.jupiter.api.Test;

public class StaticInitializerTest extends StaticInitializerTestBase {

  // The `value` will only be set if the parent class's `BeforeAll` method
  // has been called. If we've called `Class.forName(String)` in the
  // `ActualRunner` then this value will be `null`, so the test won't even
  // start. This test verifies that while we can still find tests, we're
  // not initialising classes until they're used.
  private static final int length = value.length();

  @Test
  public void doSomethingToForceInitializersToBeRun() {
    assertNotEquals(0, length);
  }
}

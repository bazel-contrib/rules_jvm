package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertThrows;

public class CallingSystemExitTest {

  @Test
  public void shouldBeAbleToCallSystemExitInATest() {
    assertThrows(SecurityException.class, () -> System.exit(2));
  }
}

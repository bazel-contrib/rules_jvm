package com.github.bazel_contrib.rules_jvm_contrib.junit5;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertThrows;

public class CallingSystemExitTest {

  @Test
  public void shouldBeAbleToCallSystemExitInATest() {
    assertThrows(SecurityException.class, () -> System.exit(2));
  }
}

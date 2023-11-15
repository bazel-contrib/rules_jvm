package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;

public abstract class ExcludedBaseTest {

  @Test
  public void defaultTestForOtherTests() {
    assertTrue(false);
  }
}

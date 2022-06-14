package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.Assumptions;
import org.junit.jupiter.api.Test;

public class AssumptionsJUnit5Test {

  @Test
  void shouldBeSkipped() {
    Assumptions.assumeTrue(false);
    System.out.println("<<<< shouldBeSkipped");
  }

  @Test
  void shouldBeExecuted() {
    System.out.println(">>>> shouldBeExecuted");
  }
}

package com.github.bazel_contrib.contrib_rules_jvm.junit4;

import org.junit.Assume;
import org.junit.Test;

public class AssumptionsJUnit4Test {

  @Test
  public void shouldBeSkipped() {
    Assume.assumeTrue(false);
    System.out.println("<<<< shouldBeSkipped");
  }

  @Test
  public void shouldBeExecuted() {
    System.out.println(">>>> shouldBeExecuted");
  }
}

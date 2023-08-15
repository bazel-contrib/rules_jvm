package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.Test;
import org.junit.experimental.runners.Enclosed;
import org.junit.runner.RunWith;

@RunWith(Enclosed.class)
public class NestedClassesVintageTest {
  public static class First {
    @Test
    public void shouldBeExecuted() {
      System.out.println(">>>> Executed test in NestedClassesVintageTest$First");
    }
  }
}

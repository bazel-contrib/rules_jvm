package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.BeforeAll;

public class StaticInitializerTestBase {

  protected static String value;

  @BeforeAll
  public static void setTestData() {
    value = "Hello, World!";
  }
}

package com.github.bazel_contrib.contrib_rules_jvm.junit5.tags;

import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;

public class Junit5TagsTest {

  @Test
  @Tag("Always")
  public void alwaysRun() {
    assertTrue(true);
  }

  @Test
  @Tag("Never")
  public void neverRun() {
    assertTrue(false);
  }

  @Test
  @Tag("Sometimes")
  public void runSometimes() {
    // exclude shouldn't run
    if (System.getProperty("JUNIT5_EXCLUDE_TAGS", "").contains("Sometimes")) {
      assertTrue(false, "Should have skippend this test");
    } else if (System.getProperty("JUNIT5_INCLUDE_TAGS", "").contains("Sometimes")) {
      assertTrue(true, "Positive ask to run this test");
    } else {
      assertTrue(true, "Not specifically filtered, but run anyway");
    }
  }
}

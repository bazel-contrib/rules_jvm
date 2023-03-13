package com.github.bazel_contrib.contrib_rules_jvm.junit5.sample;

import org.junit.jupiter.api.RepeatedTest;
import org.junit.jupiter.api.RepetitionInfo;
import org.junit.jupiter.api.Test;

public class ShardingTestTests {
  @Test
  void testOne() {}

  @Test
  void testTwo() {}

  @Test
  void testThree() {}

  @Test
  void testFour() {}

  @Test
  void testFive() {}

  @Test
  void testSix() {}

  @Test
  void testSeven() {}

  @Test
  void testEight() {}

  @RepeatedTest(value = 8)
  void testRepeated(RepetitionInfo info) {}
}

package com.github.bazel_contrib.contrib_rules_jvm.junit5.sample;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

public class ShardingTestMoreTests {
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

  @ParameterizedTest
  @ValueSource(ints = {0, 1, 2, 3, 4, 5, 6, 7})
  void testParameterized(int value) {}
}

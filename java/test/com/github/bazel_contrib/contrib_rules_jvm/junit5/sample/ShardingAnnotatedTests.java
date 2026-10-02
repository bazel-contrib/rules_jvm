package com.github.bazel_contrib.contrib_rules_jvm.junit5.sample;

import com.github.bazel_contrib.contrib_rules_jvm.junit5.ShardTemplateInvocations;
import org.junit.jupiter.api.RepeatedTest;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.ValueSource;

public class ShardingAnnotatedTests {
  @ShardTemplateInvocations
  @ParameterizedTest
  @ValueSource(ints = {0, 1, 2, 3, 4, 5, 6})
  void testShardedParameterized(int value) {}

  @ShardTemplateInvocations
  @RepeatedTest(5)
  void testShardedRepeated() {}

  @ParameterizedTest
  @ValueSource(ints = {0, 1, 2})
  void testUnshardedParameterized(int value) {}
}

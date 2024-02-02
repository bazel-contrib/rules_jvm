package com.github.bazel_contrib.contrib_rules_jvm.comparative;

import java.util.stream.Stream;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;

public class MixedTest {

  private static Stream<Arguments> argsProvider() {
    return Stream.of(Arguments.of("alpha"), Arguments.of("beta"), Arguments.of("gamma"));
  }

  @ParameterizedTest
  @MethodSource("argsProvider")
  public void bootstrap(String goGreek) {
    System.out.println(goGreek);
  }

  @Test
  void empty() {
    // Does nothing
  }
}

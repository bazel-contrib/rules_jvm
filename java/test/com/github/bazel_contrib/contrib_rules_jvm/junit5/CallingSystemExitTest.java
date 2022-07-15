package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertThrows;

import edu.umd.cs.findbugs.annotations.SuppressFBWarnings;
import org.junit.jupiter.api.Test;

@SuppressFBWarnings("THROWS_METHOD_THROWS_CLAUSE_THROWABLE")
public class CallingSystemExitTest {

  @Test
  public void shouldBeAbleToCallSystemExitInATest() {
    assertThrows(SecurityException.class, () -> System.exit(2));
  }
}

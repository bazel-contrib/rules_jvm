package com.github.bazelcontrib.rulesjvm.example;

import static org.junit.jupiter.api.Assertions.assertEquals;

import org.junit.jupiter.api.Test;

public class PassingTest {

  @Test
  public void passingText() {
    assertEquals("one", "one");
  }
}

package com.github.bazelcontrib.rulesjvm.example;

import static org.junit.jupiter.api.Assertions.assertEquals;

import org.junit.jupiter.api.Test;

public class FailingTest {

  @Test
  public void passingText() {
    assertEquals("one", "two");
  }
}

package com.example.test;

import org.junit.Test;

public class SomeTest {
  @Test
  public void testPasses() {
    if (1 != 1) {
      throw new RuntimeException("Unexpected");
    }
  }
}

package com.example.hello.notworld.withhelpers;

import static org.junit.Assert.assertEquals;

import org.junit.jupiter.api.Test;

public class NotWorldTest {
  @Test
  public void notWorld() {
    assertEquals(Helper.getExpectation(), NotWorld.NOT_WORLD);
  }
}

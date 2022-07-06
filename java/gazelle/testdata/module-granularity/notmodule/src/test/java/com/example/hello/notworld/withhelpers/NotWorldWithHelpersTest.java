package com.example.hello.hello.notworld.withhelpers;

import static org.junit.Assert.assertEquals;

import org.junit.jupiter.api.Test;

public class NotWorldTest {
  @Test
  public void notWorld() {
    assertEquals("NOT WORLD!", Helper.toUpperCase(NotWorld.NOT_WORLD));
  }
}

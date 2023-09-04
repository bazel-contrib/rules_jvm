package com.example.hello.notworld;

import static org.junit.jupiter.api.Assertions.assertEquals;

import org.junit.jupiter.api.Test;

public class NotWorldTest {
  @Test
  public void notWorld() {
    assertEquals("Not World!", NotWorld.NOT_WORLD);
  }
}

package com.example.hello.notworld.withhelpers.withdirectory;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.equalTo;

import org.junit.Test;

public class AnotherTest {
  @Test
  public void notWorld() {
    assertEquals("NOT WORLD!", Helper.toUpperCase(NotWorld.NOT_WORLD));
  }
}

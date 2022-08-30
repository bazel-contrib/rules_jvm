package com.example.myproject;

import static org.junit.Assert.assertEquals;

import org.junit.Test;

public class AppTest {

  @Test
  public void testCompare() throws Exception {
    App app = new App();
    assertEquals("should return 0 when both numbers are equal", 0, app.compare(1, 1));
  }
}

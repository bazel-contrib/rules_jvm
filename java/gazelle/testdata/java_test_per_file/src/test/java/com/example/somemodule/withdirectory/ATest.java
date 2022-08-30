package com.example.somemodule.withdirectory;

import static org.junit.Assert.assertEquals;

import com.example.myproject.App;
import org.junit.Test;

public class ATest {

  @Test
  public void testCompare() throws Exception {
    App app = new App();
    assertEquals("should return 0 when both numbers are equal", 0, app.compare(1, 1));
  }
}

package com.example.myproject;

import com.example.annotation.FlakyTest;
import org.junit.jupiter.api.Test;

import java.util.Random;

import static org.junit.jupiter.api.Assertions.assertEquals;

@FlakyTest
public class RandomTest {
  @Test
  public void unreliableTest() {
    Random random = new Random();
    int r = random.nextInt(2);
    assertEquals(r, 0);
  }
}

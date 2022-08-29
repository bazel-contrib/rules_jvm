package com.example.myproject;

import org.junit.jupiter.api.Test;

import java.util.Random;

import static org.junit.jupiter.api.Assertions.assertTrue;

public class ReliableTest {
  @Test
  public void reliableTest() {
    Random random = new Random();
    int r = random.nextInt(2);
    assertTrue(r < 2);
  }
}

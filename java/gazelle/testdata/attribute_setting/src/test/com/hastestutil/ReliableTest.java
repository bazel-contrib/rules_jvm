package com.example.hastestutil;

import org.junit.jupiter.api.Test;

import java.util.Random;

import static org.junit.jupiter.api.Assertions.assertTrue;

public class ReliableTest {
  @Test
  public void reliableTest() {
    int r = RandomNumberGenerator.generateNumberLessThanTwo();
    assertTrue(r < 2);
  }
}

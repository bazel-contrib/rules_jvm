package com.example.hastestutil;

import com.example.annotation.FlakyTest;
import org.junit.jupiter.api.Test;

import java.util.Random;

import static org.junit.jupiter.api.Assertions.assertEquals;

@FlakyTest
public class RandomTest {
  @Test
  public void unreliableTest() {
    int r = RandomNumberGenerator.generateNumberLessThanTwo();
    assertEquals(r, 0);
  }
}

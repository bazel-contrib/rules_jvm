package com.example.tests;

import com.example.annotations.RequiresNetwork;

import org.junit.jupiter.api.Test;

@RequiresNetwork
public class RequiresNetworkTest {
  @Test
  public void passes() {}
}


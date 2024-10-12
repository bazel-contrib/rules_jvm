package com.github.bazelcontrib.rulesjvm.example;

import com.google.common.base.Joiner;

public class Main {
  public static void main(String[] args) {
    System.out.println(Joiner.on(" ").join("Hello,", "World!"));
  }
}

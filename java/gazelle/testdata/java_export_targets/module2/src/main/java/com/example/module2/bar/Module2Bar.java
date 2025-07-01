package com.example.module2.bar;

public class Module2Bar {
  public static String module() {
    return "Module2Bar" + DependOnBaz.module();
  }
}

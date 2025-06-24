package com.example.module2.bar;

import com.example.module2.baz.Module2Baz;

public class DependOnBaz {
  public static String module() {
    return "Module2DependOnBaz: " + Module2Baz.module();
  }
}

package com.example.module1;

import com.example.module1.foo.Module1Foo;

public class Module1 {
  public static String module() {

    return "Module1 + " + Module1Foo.module();
  }
}

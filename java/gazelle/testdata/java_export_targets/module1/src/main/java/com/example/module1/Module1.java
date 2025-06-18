package com.example.module1;

import com.example.module1.foo.Module1Foo;

public class Module1 {
  public static String module(Module1Foo foo) {
    return "Module1 + " + foo.nonStaticModule();
  }

  public Module1Foo getFoo() {
    return new Module1Foo();
  }
}

package com.example.user;

import com.example.module1.Module1;
import com.example.module1.Module1InterfaceToFoo;
import com.example.module1.foo.Module1Foo;
import com.example.module2.bar.Module2Bar;
import com.example.module2.baz.Module2Baz;
import com.example.nested.UseChild;
import com.example.nested.child_export.ChildExport;

public class UseAllModules {
  public static void main(String[] args) {
    System.out.println(Module1.module(new Module1Foo()));
    System.out.println(Module1Foo.module());
    System.out.println(Module2Bar.module());
    System.out.println(Module2Baz.module());
    System.out.println(UseChild.module());
    System.out.println(ChildExport.module());
  }

  public void useInterfaceModule1Foo(Module1InterfaceToFoo i) {
  }
}

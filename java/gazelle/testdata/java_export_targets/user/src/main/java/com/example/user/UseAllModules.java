package com.example.user;

import com.example.module1.Module1;
import com.example.module1.foo.Module1Foo;
import com.example.module2.bar.Module2Bar;
import com.example.module2.baz.Module2Baz;
import com.example.nested.UseChild;
import com.example.nested.child_export.ChildExport;
import com.example.export_depending_on_different_package.export.DependOnLib;
import com.example.other_deps.runtime.RuntimeDep;
import com.example.other_deps.plain_deps.PlainDep;
import com.example.other_deps.third_party.ThirdPartyDeps;
import com.example.shared_dep.Dependent;

public class UseAllModules {
  public static void main(String[] args) {
    System.out.println(Module1.module(new Module1Foo()));
    System.out.println(Module1Foo.module());
    System.out.println(Module2Bar.module());
    System.out.println(Module2Baz.module());
    System.out.println(UseChild.module());
    System.out.println(ChildExport.module());
    System.out.println(DependOnLib.module());
    System.out.println(PlainDep.module());
    System.out.println(RuntimeDep.module());
    System.out.println(ThirdPartyDeps.module());
    System.out.println(Dependent.module());
  }
}

package com.example.export_depending_on_different_package.export;

import com.example.export_depending_on_different_package.lib.Lib;

public class DependOnLib {
  public static String module() {
    return "DependOnLib + " + Lib.module();
  }
}

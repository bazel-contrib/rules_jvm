package com.example.other_deps.plain_deps;

import com.example.other_deps.plain_deps.dep.PlainDep;

public class PlainSrc {
  public static String module() {
    return "PlainSrc + " + PlainDep.module();
  }
}

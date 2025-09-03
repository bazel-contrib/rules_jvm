package com.example.shared_dep;

import com.example.shared_deps.shared_dep.SharedDep;

public class Dependent {
  public static String module() {
    return "Dependent +" + SharedDep.module();
  }
}

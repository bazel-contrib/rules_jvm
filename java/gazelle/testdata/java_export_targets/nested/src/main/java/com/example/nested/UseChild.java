package com.example.nested;

public class NestedExport {
  public static String module() {
    return "Parent of " + ChildExport.module();
  }
}

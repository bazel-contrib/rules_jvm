package com.example.nested;

import com.example.nested.child_export.ChildExport;

public class NestedExport {
  public static String module() {
    return "Parent of " + ChildExport.module();
  }
}

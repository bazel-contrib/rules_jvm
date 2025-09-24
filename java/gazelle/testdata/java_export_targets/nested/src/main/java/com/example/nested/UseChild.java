package com.example.nested;

import com.example.nested.child_export.ChildExport;

public class UseChild {
  public static String module() {
    return "Parent of " + ChildExport.module();
  }
}

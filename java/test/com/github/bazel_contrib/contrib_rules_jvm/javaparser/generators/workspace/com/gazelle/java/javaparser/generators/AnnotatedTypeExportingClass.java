package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators.workspace.com.gazelle.java.javaparser.generators;

import example.external.ArrayElement;
import example.external.Outer;
import example.external.TypeArgument;
import example.external.TypeUse;
import java.util.List;

public class AnnotatedTypeExportingClass {
  public Outer.@TypeUse Inner nested() {
    return null;
  }

  public ArrayElement @TypeUse [] array() {
    return null;
  }

  public List<@TypeUse TypeArgument> typeArgument() {
    return null;
  }

  public example.external.@TypeUse FullyQualified fullyQualified() {
    return null;
  }
}

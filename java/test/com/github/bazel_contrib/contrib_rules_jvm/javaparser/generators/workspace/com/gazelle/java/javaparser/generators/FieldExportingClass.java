package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators.workspace.com.gazelle.java.javaparser.generators;

import example.external.AnonymousClassField;
import example.external.LocalVariable;
import example.external.PackageField;
import example.external.ParameterizedFieldArg;
import example.external.ParameterizedFieldOuter;
import example.external.PrivateField;
import example.external.ProtectedField;
import example.external.PublicField;

public class FieldExportingClass {
  public PublicField publicField;
  protected ProtectedField protectedField;
  PackageField packageField;
  private PrivateField privateField;
  public ParameterizedFieldOuter<ParameterizedFieldArg> parameterizedField;

  public Runnable runnable =
      new Runnable() {
        public AnonymousClassField hidden = null;

        @Override
        public void run() {}
      };

  public void method() {
    LocalVariable local = null;
  }
}

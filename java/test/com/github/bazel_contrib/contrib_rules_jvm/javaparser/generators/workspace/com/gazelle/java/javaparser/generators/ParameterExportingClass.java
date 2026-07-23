package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators.workspace.com.gazelle.java.javaparser.generators;

import example.external.ConstructorParam;
import example.external.PackageParam;
import example.external.ParameterizedArg;
import example.external.ParameterizedParam;
import example.external.PrivateParam;
import example.external.ProtectedParam;
import example.external.PublicParam;
import example.external.VarargParam;
import example.external.WildcardBound;

public class ParameterExportingClass {
  public ParameterExportingClass(ConstructorParam param) {}

  void acceptsPackage(PackageParam param) {}

  private void acceptsPrivate(PrivateParam param) {}

  protected void acceptsProtected(ProtectedParam param) {}

  public void acceptsPublic(PublicParam param) {}

  public void acceptsVarargs(VarargParam... params) {}

  public void acceptsParameterized(ParameterizedParam<ParameterizedArg> param) {}

  public void acceptsWildcard(ParameterizedParam<? extends WildcardBound> param) {}
}

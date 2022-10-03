package com.github.bazel_contrib.contrib_rules_jvm.junit5;

public class NullSystemExitToggle implements SystemExitToggle {
  @Override
  public void prevent() {
    // No-op
  }

  @Override
  public void allow() {
    // No-op
  }
}

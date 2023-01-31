package com.github.bazel_contrib.contrib_rules_jvm.junit5;

public interface SystemExitToggle {

  void prevent();

  void allow();
}

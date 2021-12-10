package com.github.bazel_contrib.contrib_rules_jvm.junit5;

@FunctionalInterface
public interface RunsTest {
  boolean run(String testClassName);
}

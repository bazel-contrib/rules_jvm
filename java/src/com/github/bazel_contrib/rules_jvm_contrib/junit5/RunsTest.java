package com.github.bazel_contrib.rules_jvm_contrib.junit5;

@FunctionalInterface
public interface RunsTest {
  boolean run(String testClassName);
}

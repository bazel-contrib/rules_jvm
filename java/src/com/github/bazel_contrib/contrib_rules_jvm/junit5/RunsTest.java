package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.util.List;

@FunctionalInterface
public interface RunsTest {
  boolean run(String testClassName, List<String> includeEngines, List<String> excludeEngines);
}

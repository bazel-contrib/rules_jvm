package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.util.concurrent.atomic.AtomicBoolean;
import org.junit.jupiter.api.extension.ConditionEvaluationResult;
import org.junit.jupiter.api.extension.ExecutionCondition;
import org.junit.jupiter.api.extension.ExtensionContext;
import org.junit.platform.engine.TestExecutionResult;
import org.junit.platform.engine.TestExecutionResult.Status;
import org.junit.platform.launcher.TestExecutionListener;
import org.junit.platform.launcher.TestIdentifier;

public class FailFastExtension implements ExecutionCondition, TestExecutionListener {
  /**
   * This environment variable is set to 1 if Bazel is run with --test_runner_fail_fast, indicating
   * that the test runner should exit as soon as possible after the first failure. {@see
   * https://github.com/bazelbuild/bazel/commit/957554037ced26dc1860b9c23445a8ccc44f697e}
   */
  private static final boolean SHOULD_FAIL_FAST =
      "1".equals(System.getenv("TESTBRIDGE_TEST_RUNNER_FAIL_FAST"));

  private static final AtomicBoolean SOME_TEST_FAILED = new AtomicBoolean();

  @Override
  public ConditionEvaluationResult evaluateExecutionCondition(ExtensionContext extensionContext) {
    if (!SHOULD_FAIL_FAST) {
      return ConditionEvaluationResult.enabled(
          "Running test since --test_runner_fail_fast is not enabled");
    }
    if (SOME_TEST_FAILED.get()) {
      return ConditionEvaluationResult.disabled(
          "Skipping test since --test_runner_fail_fast is enabled and another test has failed");
    } else {
      return ConditionEvaluationResult.enabled("Running test since no other test has failed yet");
    }
  }

  @Override
  public void executionFinished(
      TestIdentifier testIdentifier, TestExecutionResult testExecutionResult) {
    if (SHOULD_FAIL_FAST && testExecutionResult.getStatus().equals(Status.FAILED)) {
      SOME_TEST_FAILED.set(true);
    }
  }
}

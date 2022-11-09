package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;
import org.junit.platform.engine.TestDescriptor;
import org.junit.platform.engine.TestExecutionResult;
import org.junit.platform.engine.UniqueId;
import org.junit.platform.launcher.TestIdentifier;
import org.junit.platform.launcher.TestPlan;
import org.mockito.Mockito;
import org.opentest4j.TestAbortedException;

public class JunitOutputXmlTest {

  private TestDescriptor testDescriptor = new StubbedTestDescriptor(createId("descriptors"));
  private TestIdentifier identifier = TestIdentifier.from(testDescriptor);
  ;

  @Test
  public void testResultCanBeDisabled() {
    TestPlan testPlan = Mockito.mock(TestPlan.class);

    // Note: we do not call `markFinished` so the TestResult is null. This is what happens when
    // run by JUnit5.
    TestResult result = new TestResult(testPlan, identifier, false);

    assertTrue(result.isDisabled());
  }

  @Test
  public void disabledTestsAreNotSkipped() {
    TestPlan testPlan = Mockito.mock(TestPlan.class);

    // Note: we do not call `markFinished` so the TestResult is null. This is what happens when
    // run by JUnit5.
    TestResult result = new TestResult(testPlan, identifier, false);

    assertFalse(result.isSkipped());
  }

  @Test
  public void skippedTestsAreMarkedAsSkipped() {
    TestPlan testPlan = Mockito.mock(TestPlan.class);

    TestResult result = new TestResult(testPlan, identifier, false);
    TestExecutionResult testExecutionResult =
        TestExecutionResult.aborted(new TestAbortedException("skipping is fun"));
    result.markFinished(testExecutionResult);

    assertTrue(result.isSkipped());
  }

  @Test
  public void skippedTestsAreNotDisabled() {
    TestPlan testPlan = Mockito.mock(TestPlan.class);

    TestResult result = new TestResult(testPlan, identifier, false);
    TestExecutionResult testExecutionResult =
        TestExecutionResult.aborted(new TestAbortedException("skipping is fun"));
    result.markFinished(testExecutionResult);

    assertFalse(result.isDisabled());
  }

  private UniqueId createId(String testName) {
    return UniqueId.parse(
        String.format("[engine:mocked]/[class:ExampleTest]/[method:%s()]", testName));
  }
}

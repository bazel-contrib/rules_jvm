package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.util.LinkedList;
import java.util.List;
import java.util.Optional;
import javax.xml.stream.XMLStreamWriter;
import org.junit.platform.engine.TestExecutionResult;
import org.junit.platform.engine.reporting.ReportEntry;
import org.junit.platform.launcher.TestIdentifier;
import org.junit.platform.launcher.TestPlan;

public class RootContainer extends BaseResult {

  // Insertion order matters when we come to output the results
  private final List<TestSuiteResult> suites = new LinkedList<>();
  private final TestPlan testPlan;

  public RootContainer(TestIdentifier rootId, TestPlan testPlan) {
    super(rootId);
    this.testPlan = testPlan;

    testPlan.getChildren(rootId).forEach(child -> suites.add(createSuite(child)));
  }

  public void addDynamicTest(TestIdentifier testIdentifier) {
    testPlan
        .getParent(testIdentifier)
        .flatMap(this::get)
        .filter(TestSuiteResult.class::isInstance)
        .map(TestSuiteResult.class::cast)
        .ifPresent(
            suite -> suite.add(new TestResult(testPlan, testIdentifier, /*isDynamic=*/ true)));
  }

  public void markStarted(TestIdentifier testIdentifier) {
    get(testIdentifier).ifPresent(BaseResult::markStarted);
  }

  public void markSkipped(TestIdentifier testIdentifier, String reason) {
    get(testIdentifier).ifPresent(result -> result.markSkipped(reason));
  }

  public void markFinished(TestIdentifier testIdentifier, TestExecutionResult testExecutionResult) {
    get(testIdentifier).ifPresent(result -> result.markFinished(testExecutionResult));
  }

  public void addReportingEntry(TestIdentifier testIdentifier, ReportEntry entry) {
    get(testIdentifier).ifPresent(result -> result.addReportEntry(entry));
  }

  protected Optional<BaseResult> get(TestIdentifier testIdentifier) {
    if (getTestId().equals(testIdentifier)) {
      return Optional.of(this);
    }

    return suites.stream()
        .map(suite -> suite.get(testIdentifier))
        .filter(Optional::isPresent)
        .map(Optional::get)
        .findFirst();
  }

  public void toXml(XMLStreamWriter xml) {
    write(() -> suites.forEach(suite -> suite.toXml(xml)));
  }

  private TestSuiteResult createSuite(TestIdentifier suiteId) {
    TestSuiteResult suite = new TestSuiteResult(suiteId);
    for (TestIdentifier child : testPlan.getChildren(suiteId)) {
      if (child.isContainer()) {
        suite.add(createSuite(child));
      } else {
        suite.add(new TestResult(testPlan, child, /*isDynamic=*/ false));
      }
    }
    return suite;
  }
}

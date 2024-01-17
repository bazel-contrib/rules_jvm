package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static java.nio.charset.StandardCharsets.UTF_8;

import java.io.BufferedWriter;
import java.io.Closeable;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.logging.Level;
import java.util.logging.Logger;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import javax.xml.stream.XMLOutputFactory;
import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamWriter;
import org.junit.platform.engine.TestExecutionResult;
import org.junit.platform.engine.UniqueId;
import org.junit.platform.engine.reporting.ReportEntry;
import org.junit.platform.launcher.TestExecutionListener;
import org.junit.platform.launcher.TestIdentifier;
import org.junit.platform.launcher.TestPlan;

public class BazelJUnitOutputListener implements TestExecutionListener, Closeable {
  public static final Logger LOG = Logger.getLogger(BazelJUnitOutputListener.class.getName());
  private final XMLStreamWriter xml;

  private final Object resultsLock = new Object();
  // Commented out to avoid adding a dependency to building the test runner.
  // This is really just documentation until someone actually turns on a static analyser.
  // If they do, we can decide whether we want to pick up the dependency.
  // @GuardedBy("resultsLock")
  private final Map<UniqueId, TestData> results = new ConcurrentHashMap<>();
  private TestPlan testPlan;

  // If we have already closed this listener, we shouldn't write any more XML.
  private final AtomicBoolean hasClosed = new AtomicBoolean();
  // Whether test-running was interrupted (e.g. because our tests timed out and we got SIGTERM'd)
  // and when writing results we want to flush any pending tests as interrupted,
  // rather than ignoring them because they're incomplete.
  private final AtomicBoolean wasInterrupted = new AtomicBoolean();

  public BazelJUnitOutputListener(Path xmlOut) {
    try {
      Files.createDirectories(xmlOut.getParent());
      BufferedWriter writer = Files.newBufferedWriter(xmlOut, UTF_8);
      xml = XMLOutputFactory.newFactory().createXMLStreamWriter(writer);
      xml.writeStartDocument("UTF-8", "1.0");
    } catch (IOException | XMLStreamException e) {
      throw new IllegalStateException("Unable to create output file", e);
    }
  }

  @Override
  public void testPlanExecutionStarted(TestPlan testPlan) {
    this.testPlan = testPlan;

    try {
      // Closed when we call `testPlanExecutionFinished`
      xml.writeStartElement("testsuites");
    } catch (XMLStreamException e) {
      throw new RuntimeException(e);
    }
  }

  @Override
  public void testPlanExecutionFinished(TestPlan testPlan) {
    if (this.testPlan == null) {
      throw new IllegalStateException("Test plan is not currently executing");
    }

    try {
      // Closing `testsuites` element
      xml.writeEndElement();
    } catch (XMLStreamException e) {
      throw new RuntimeException(e);
    }

    this.testPlan = null;
  }

  // Requires the caller to have acquired resultsLock.
  // Commented out to avoid adding a dependency to building the test runner.
  // This is really just documentation until someone actually turns on a static analyser.
  // If they do, we can decide whether we want to pick up the dependency.
  // @GuardedBy("resultsLock")
  private Map<TestData, List<TestData>> matchTestCasesToSuites_locked(
      List<TestData> testCases, boolean includeIncompleteTests) {
    Map<TestData, List<TestData>> knownSuites = new HashMap<>();

    // Find the containing test suites for the test cases.
    for (TestData testCase : testCases) {
      TestData parent = testCase.getId().getParentIdObject().map(results::get).orElse(null);
      if (parent == null) {
        // Something has gone wildly wrong. I really hope people file some
        // bugs with us if they run into this....
        LOG.warning("Unable to find parent test for " + testCase.getId());
        throw new IllegalStateException("Unable to find parent test for " + testCase.getId());
      }
      if (includeIncompleteTests || testCase.getDuration() != null) {
        knownSuites.computeIfAbsent(parent, id -> new ArrayList<>()).add(testCase);
      }
    }

    return knownSuites;
  }

  // Requires the caller to have acquired resultsLock.
  // Commented out to avoid adding a dependency to building the test runner.
  // This is really just documentation until someone actually turns on a static analyser.
  // If they do, we can decide whether we want to pick up the dependency.
  // @GuardedBy("resultsLock")
  private List<TestData> findTestCases_locked() {
    return results.values().stream()
        // Ignore test plan roots. These are always the engine being used.
        .filter(result -> !testPlan.getRoots().contains(result.getId()))
        .filter(
            result -> {
              // Find the test results we will convert to `testcase` entries. These
              // are identified by the fact that they have no child test cases in the
              // test plan, or they are marked as tests.
              TestIdentifier id = result.getId();
              return id.isTest() || testPlan.getChildren(id).isEmpty();
            })
        .collect(Collectors.toList());
  }

  @Override
  public void dynamicTestRegistered(TestIdentifier testIdentifier) {
    getResult(testIdentifier).setDynamic(true);
  }

  @Override
  public void executionSkipped(TestIdentifier testIdentifier, String reason) {
    getResult(testIdentifier).mark(TestExecutionResult.aborted(null)).skipReason(reason);
    outputIfTestRootIsComplete(testIdentifier);
  }

  @Override
  public void executionStarted(TestIdentifier testIdentifier) {
    getResult(testIdentifier).started();
  }

  @Override
  public void executionFinished(
      TestIdentifier testIdentifier, TestExecutionResult testExecutionResult) {
    getResult(testIdentifier).mark(testExecutionResult);
    outputIfTestRootIsComplete(testIdentifier);
  }

  private void outputIfTestRootIsComplete(TestIdentifier testIdentifier) {
    if (!testPlan.getRoots().contains(testIdentifier)) {
      return;
    }

    output(false);
  }

  private void output(boolean includeIncompleteTests) {
    synchronized (this.resultsLock) {
      List<TestData> testCases = findTestCases_locked();
      Map<TestData, List<TestData>> testSuites =
          matchTestCasesToSuites_locked(testCases, includeIncompleteTests);

      // Write the results
      try {
        for (Map.Entry<TestData, List<TestData>> suiteAndTests : testSuites.entrySet()) {
          new TestSuiteXmlRenderer(testPlan)
              .toXml(xml, suiteAndTests.getKey(), suiteAndTests.getValue());
        }
      } catch (XMLStreamException e) {
        throw new RuntimeException(e);
      }

      // Delete the results we've used to conserve memory. This is safe to do
      // since we only do this when the test root is complete, so we know that
      // we won't be adding to the list of suites and test cases for that root
      // (because tests and containers are arranged in a hierarchy --- the
      // containers only complete when all the things they contain are
      // finished. We are leaving all the test data that we have _not_ written
      // to the XML file.
      Stream.concat(testCases.stream(), testSuites.keySet().stream())
          .forEach(data -> results.remove(data.getId().getUniqueIdObject()));
    }
  }

  @Override
  public void reportingEntryPublished(TestIdentifier testIdentifier, ReportEntry entry) {
    getResult(testIdentifier).addReportEntry(entry);
  }

  private TestData getResult(TestIdentifier id) {
    synchronized (resultsLock) {
      return results.computeIfAbsent(id.getUniqueIdObject(), ignored -> new TestData(id));
    }
  }

  public void closeForInterrupt() {
    wasInterrupted.set(true);
    close();
  }

  public void close() {
    if (hasClosed.getAndSet(true)) {
      return;
    }
    if (wasInterrupted.get()) {
      output(true);
    }
    try {
      xml.writeEndDocument();
      xml.close();
    } catch (XMLStreamException e) {
      LOG.log(Level.SEVERE, "Unable to close xml output", e);
    }
  }
}

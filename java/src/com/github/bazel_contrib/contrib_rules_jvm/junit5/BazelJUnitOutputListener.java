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
  private final XMLStreamWriter xml;

  private final Map<UniqueId, TestData> results = new ConcurrentHashMap<>();
  private TestPlan testPlan;

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
      // Closed when we call `testPlanExecutionFinished
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

  private Map<TestData, List<TestData>> matchTestCasesToSuites(List<TestData> testCases) {
    Map<TestData, List<TestData>> knownSuites = new HashMap<>();

    // Find the containing test suites for the test cases.
    for (TestData testCase : testCases) {
      TestData parent = testCase.getId().getParentIdObject().map(results::get).orElse(null);
      if (parent == null) {
        // We should really log this, because something has gone wildly wrong
        continue;
      }

      knownSuites.computeIfAbsent(parent, id -> new ArrayList<>()).add(testCase);
    }

    return knownSuites;
  }

  private List<TestData> findTestCases() {
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

    List<TestData> testCases = findTestCases();
    Map<TestData, List<TestData>> testSuites = matchTestCasesToSuites(testCases);

    // Write the results
    try {
      for (Map.Entry<TestData, List<TestData>> suiteAndTests : testSuites.entrySet()) {
        new TestSuiteXmlRenderer(testPlan)
            .toXml(xml, suiteAndTests.getKey(), suiteAndTests.getValue());
      }
    } catch (XMLStreamException e) {
      throw new RuntimeException(e);
    }

    // Delete the results we've used to conserve memory
    Stream.concat(testCases.stream(), testSuites.keySet().stream())
        .forEach(data -> results.remove(data.getId().getUniqueIdObject()));
  }

  @Override
  public void reportingEntryPublished(TestIdentifier testIdentifier, ReportEntry entry) {
    getResult(testIdentifier).addReportEntry(entry);
  }

  private TestData getResult(TestIdentifier id) {
    return results.computeIfAbsent(id.getUniqueIdObject(), ignored -> new TestData(id));
  }

  public void close() {
    try {
      xml.writeEndDocument();
      xml.close();
    } catch (XMLStreamException e) {
      // LOG.log(Level.SEVERE, "Unable to close xml output", e);
    }
  }
}

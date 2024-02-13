package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assertions.fail;
import static org.junit.platform.launcher.LauncherConstants.STDERR_REPORT_ENTRY_KEY;
import static org.junit.platform.launcher.LauncherConstants.STDOUT_REPORT_ENTRY_KEY;

import java.io.IOException;
import java.io.Reader;
import java.io.StringReader;
import java.io.StringWriter;
import java.io.Writer;
import java.util.Collection;
import java.util.List;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import javax.xml.parsers.ParserConfigurationException;
import javax.xml.stream.XMLOutputFactory;
import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamWriter;
import org.junit.jupiter.api.Test;
import org.junit.platform.engine.TestDescriptor;
import org.junit.platform.engine.TestExecutionResult;
import org.junit.platform.engine.UniqueId;
import org.junit.platform.engine.reporting.ReportEntry;
import org.junit.platform.launcher.TestIdentifier;
import org.junit.platform.launcher.TestPlan;
import org.mockito.Mockito;
import org.opentest4j.TestAbortedException;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.w3c.dom.Node;
import org.w3c.dom.NodeList;
import org.xml.sax.InputSource;
import org.xml.sax.SAXException;

public class BazelJUnitOuputListenerTest {

  private TestDescriptor testDescriptor = new StubbedTestDescriptor(createId("descriptors"));
  private TestIdentifier identifier = TestIdentifier.from(testDescriptor);

  @Test
  public void testResultCanBeDisabled() {
    // Note: we do not call `markFinished` so the TestResult is null. This is what happens when
    // run by JUnit5.
    TestData result = new TestData(identifier);

    assertTrue(result.isDisabled());
  }

  @Test
  public void disabledTestsAreNotSkipped() {
    // Note: we do not call `markFinished` so the TestResult is null. This is what happens when
    // run by JUnit5.
    TestData result = new TestData(identifier);

    assertFalse(result.isSkipped());
  }

  @Test
  public void skippedTestsAreMarkedAsSkipped() {
    TestData result =
        new TestData(identifier)
            .mark(TestExecutionResult.aborted(new TestAbortedException("skipping is fun")));

    assertTrue(result.isSkipped());
  }

  @Test
  public void skippedTestsAreNotDisabled() {
    TestData result =
        new TestData(identifier)
            .mark(TestExecutionResult.aborted(new TestAbortedException("skipping is fun")));

    assertFalse(result.isDisabled());
    assertTrue(result.isSkipped());
  }

  @Test
  public void skippedTestsAreNotFailures() {
    TestData result =
        new TestData(identifier)
            .mark(TestExecutionResult.aborted(new TestAbortedException("skipping is fun")));

    assertTrue(result.isSkipped());
    assertFalse(result.isFailure());
    assertFalse(result.isError());
    assertFalse(result.isDisabled());
  }

  @Test
  public void skippedTestsContainMessages() {
    TestData result =
        new TestData(identifier)
            .mark(TestExecutionResult.aborted(new TestAbortedException("skipping is fun")));

    TestPlan testPlan = Mockito.mock(TestPlan.class);
    var root = generateTestXml(testPlan, result).getDocumentElement();
    assertNotNull(root);
    assertEquals("testcase", root.getTagName());

    var skipped = root.getElementsByTagName("skipped");
    assertEquals(1, skipped.getLength());

    var failures = root.getElementsByTagName("failure");
    assertEquals(0, failures.getLength());

    var message = skipped.item(0).getFirstChild();
    assertNotNull(message);
    assertEquals("skipping is fun", message.getTextContent());
  }

  @Test
  public void disabledTestsAreMarkedAsSkipped() {
    TestData suite = new TestData(identifier).skipReason("Not today");

    TestIdentifier childId = TestIdentifier.from(new StubbedTestDescriptor(createId("child")));
    TestPlan testPlan = Mockito.mock(TestPlan.class);

    TestData childResult =
        new TestData(childId)
            .mark(TestExecutionResult.aborted(new TestAbortedException("skipping is fun")));

    Document xml = generateTestXml(testPlan, suite, List.of(childResult));

    // Because of the way we generated the XML, the root element is the `testsuite` one
    Element root = xml.getDocumentElement();
    assertEquals("testsuite", root.getTagName());
    assertEquals("1", root.getAttribute("tests"));
    assertEquals("0", root.getAttribute("disabled"));

    NodeList allTestCases = root.getElementsByTagName("testcase");
    assertEquals(1, allTestCases.getLength());
    Node testCase = allTestCases.item(0);

    Node skipped = getChild("skipped", testCase);

    assertNotNull(skipped);
  }

  @Test
  public void interruptedTestsAreMarkedAsFailed() {
    TestData suite = new TestData(identifier);

    TestPlan testPlan = Mockito.mock(TestPlan.class);

    TestIdentifier completedChild =
        TestIdentifier.from(new StubbedTestDescriptor(createId("complete-child")));
    TestData completedChildResult = new TestData(completedChild).started();
    completedChildResult.mark(TestExecutionResult.successful());

    TestIdentifier interruptedChild =
        TestIdentifier.from(new StubbedTestDescriptor(createId("interrupted-child")));
    TestData interruptedChildResult = new TestData(interruptedChild).started();

    Document xml =
        generateTestXml(testPlan, suite, List.of(completedChildResult, interruptedChildResult));

    // Because of the way we generated the XML, the root element is the `testsuite` one
    Element root = xml.getDocumentElement();
    assertEquals("testsuite", root.getTagName());
    assertEquals("2", root.getAttribute("tests"));
    assertEquals("0", root.getAttribute("disabled"));
    assertEquals("0", root.getAttribute("errors"));
    assertEquals("0", root.getAttribute("skipped"));
    assertEquals("1", root.getAttribute("failures"));

    NodeList allTestCases = root.getElementsByTagName("testcase");
    assertEquals(2, allTestCases.getLength());
    Node testCaseZero = allTestCases.item(0);
    Node testCaseOne = allTestCases.item(1);

    Node failureZero = getChild("failure", testCaseZero);
    Node failureOne = getChild("failure", testCaseOne);

    if (!(failureZero == null ^ failureOne == null)) {
      fail(
          String.format("Expected exactly one failure but got %s and %s", failureZero, failureOne));
    }

    Node failure = failureZero == null ? failureOne : failureZero;

    assertEquals("Test timed out and was interrupted", failure.getTextContent());
  }

  @Test
  void throwablesWithNullMessageAreSerialized() {
    var test = new TestData(identifier).mark(TestExecutionResult.failed(new Throwable()));

    var root = generateTestXml(Mockito.mock(TestPlan.class), test).getDocumentElement();
    assertNotNull(root);
    assertEquals("testcase", root.getTagName());

    var failures = root.getElementsByTagName("failure");
    assertEquals(1, failures.getLength());

    var message = failures.item(0).getAttributes().getNamedItem("message");
    assertNotNull(message);
    assertEquals("null", message.getTextContent());
  }

  @Test
  void throwablesWithIllegalXmlCharactersInMessageAreSerialized() {
    var test =
        new TestData(identifier)
            .mark(
                TestExecutionResult.failed(
                    new Throwable(
                        "legal: \u0009"
                            + " | \n" // #xA
                            + " | \r" // #xD
                            + " | [\u0020-\uD7FF]"
                            + " | [\uE000-\uFFFD]"
                            + ", illegal: [\0-\u0008]"
                            + " | [\u000B-\u000C]"
                            + " | [\u000E-\u0019]"
                            + " | [\uD800-\uDFFF]"
                            + " | [\uFFFE-\uFFFF]")));

    var root = generateTestXml(Mockito.mock(TestPlan.class), test).getDocumentElement();
    assertNotNull(root);
    assertEquals("testcase", root.getTagName());

    var failures = root.getElementsByTagName("failure");
    assertEquals(1, failures.getLength());

    var message = failures.item(0).getAttributes().getNamedItem("message");
    assertNotNull(message);
    assertEquals(
        "legal:   |   |   | [ -\uD7FF] | [\uE000-�]"
            + ", illegal: [&#0;-&#8;]"
            + " | [&#11;-&#12;]"
            + " | [&#14;-&#25;]"
            + " | [&#55296;-&#57343;]"
            + " | [&#65534;-&#65535;]",
        message.getTextContent());
  }

  @Test
  public void ensureOutputsAreProperlyEscaped() {
    var test = new TestData(identifier);
    test.addReportEntry(ReportEntry.from(STDOUT_REPORT_ENTRY_KEY, "\u001B[31moh noes!\u001B[0m"));
    test.addReportEntry(ReportEntry.from(STDERR_REPORT_ENTRY_KEY, "\u001B[31mAlso bad!\u001B[0m"));
    test.mark(TestExecutionResult.successful());

    Document xml = generateTestXml(Mockito.mock(TestPlan.class), test);

    Node item = xml.getElementsByTagName("system-out").item(0);
    Node cdata = item.getFirstChild();
    assertEquals(Node.CDATA_SECTION_NODE, cdata.getNodeType());
    String text = cdata.getTextContent();
    // The escape characters should have been (uh) escaped.
    assertEquals("&#27;[31moh noes!&#27;[0m", text);

    item = xml.getElementsByTagName("system-err").item(0);
    cdata = item.getFirstChild();
    assertEquals(Node.CDATA_SECTION_NODE, cdata.getNodeType());
    text = cdata.getTextContent();
    // The escape characters should have been (uh) escaped.
    assertEquals("&#27;[31mAlso bad!&#27;[0m", text);
  }

  private Document generateTestXml(TestPlan testPlan, TestData testCase) {
    return generateDocument(xml -> new TestCaseXmlRenderer(testPlan).toXml(xml, testCase));
  }

  private Document generateTestXml(
      TestPlan testPlan, TestData suite, Collection<TestData> testCases) {
    return generateDocument(xml -> new TestSuiteXmlRenderer(testPlan).toXml(xml, suite, testCases));
  }

  private Document generateDocument(XmlGenerator renderer) {
    try {
      Writer writer = new StringWriter();
      XMLStreamWriter xsw = XMLOutputFactory.newDefaultFactory().createXMLStreamWriter(writer);

      renderer.accept(xsw);

      DocumentBuilderFactory factory = DocumentBuilderFactory.newInstance();
      DocumentBuilder builder;
      Reader reader = new StringReader(writer.toString());
      builder = factory.newDocumentBuilder();
      return builder.parse(new InputSource(reader));
    } catch (XMLStreamException | ParserConfigurationException | SAXException | IOException e) {
      fail(e.getMessage());
      return null; // We never get here
    }
  }

  private Node getChild(String childTagName, Node withinNode) {
    NodeList childNodes = withinNode.getChildNodes();
    for (int i = 0; i < childNodes.getLength(); i++) {
      Node node = childNodes.item(i);
      if (node.getNodeType() != Node.ELEMENT_NODE) {
        continue;
      }
      if (childTagName.equals(node.getNodeName())) {
        return node;
      }
    }
    return null;
  }

  private UniqueId createId(String testName) {
    return UniqueId.parse(
        String.format("[engine:mocked]/[class:ExampleTest]/[method:%s()]", testName));
  }

  private interface XmlGenerator {
    void accept(XMLStreamWriter xml) throws XMLStreamException;
  }
}

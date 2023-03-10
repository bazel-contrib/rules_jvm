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

public class JunitOutputXmlTest {

  private TestDescriptor testDescriptor = new StubbedTestDescriptor(createId("descriptors"));
  private TestIdentifier identifier = TestIdentifier.from(testDescriptor);

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

  @Test
  public void disabledTestsAreMarkedAsSkipped() {
    TestSuiteResult suite = new TestSuiteResult(identifier);
    suite.markSkipped("Not today");

    TestIdentifier childId = TestIdentifier.from(new StubbedTestDescriptor(createId("child")));
    TestPlan testPlan = Mockito.mock(TestPlan.class);

    TestResult childResult = new TestResult(testPlan, childId, false);
    TestExecutionResult testExecutionResult =
        TestExecutionResult.aborted(new TestAbortedException("skipping is fun"));
    childResult.markFinished(testExecutionResult);

    suite.add(new TestResult(testPlan, childId, false));

    Document xml = generateTestXml(suite);

    // Because of the way we generated the XML, the root element is the `testsuite` one
    Element root = xml.getDocumentElement();
    assertEquals("testsuite", root.getTagName());
    assertEquals("1", root.getAttribute("tests"));
    assertEquals("1", root.getAttribute("disabled"));

    NodeList allTestCases = root.getElementsByTagName("testcase");
    assertEquals(1, allTestCases.getLength());
    Node testCase = allTestCases.item(0);

    boolean skippedSeen = containsChild("skipped", testCase);

    assertTrue(skippedSeen);
  }

  @Test
  void throwablesWithNullMessageAreSerialized() {
    var test = new TestResult(Mockito.mock(TestPlan.class), identifier, false);
    test.markFinished(TestExecutionResult.failed(new Throwable()));

    var root = generateTestXml(test).getDocumentElement();
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
    var test = new TestResult(Mockito.mock(TestPlan.class), identifier, false);
    test.markFinished(
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

    var root = generateTestXml(test).getDocumentElement();
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
    var test = new TestResult(Mockito.mock(TestPlan.class), identifier, false);
    test.addReportEntry(ReportEntry.from(STDOUT_REPORT_ENTRY_KEY, "\u001B[31moh noes!\u001B[0m"));
    test.addReportEntry(ReportEntry.from(STDERR_REPORT_ENTRY_KEY, "\u001B[31mAlso bad!\u001B[0m"));
    test.markFinished(TestExecutionResult.successful());

    Document xml = generateTestXml(test);

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

  private Document generateTestXml(BaseResult result) {
    try {
      Writer writer = new StringWriter();
      XMLStreamWriter xsw = XMLOutputFactory.newDefaultFactory().createXMLStreamWriter(writer);
      result.toXml(xsw);

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

  private boolean containsChild(String childTagName, Node withinNode) {
    NodeList childNodes = withinNode.getChildNodes();
    for (int i = 0; i < childNodes.getLength(); i++) {
      Node node = childNodes.item(i);
      if (node.getNodeType() != Node.ELEMENT_NODE) {
        continue;
      }
      if (childTagName.equals(node.getNodeName())) {
        return true;
      }
    }
    return false;
  }

  private UniqueId createId(String testName) {
    return UniqueId.parse(
        String.format("[engine:mocked]/[class:ExampleTest]/[method:%s()]", testName));
  }
}

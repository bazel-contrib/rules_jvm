package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static com.github.bazel_contrib.contrib_rules_jvm.junit5.SafeXml.writeTextElement;

import java.util.Collection;
import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamWriter;
import org.junit.platform.launcher.TestPlan;

class TestSuiteXmlRenderer {

  private final TestCaseXmlRenderer testRenderer;

  public TestSuiteXmlRenderer(TestPlan testPlan) {
    testRenderer = new TestCaseXmlRenderer(testPlan);
  }

  public void toXml(XMLStreamWriter xml, TestData suite, Collection<TestData> tests)
      throws XMLStreamException {
    xml.writeStartElement("testsuite");

    xml.writeAttribute("name", suite.getId().getLegacyReportingName());
    xml.writeAttribute("tests", String.valueOf(tests.size()));

    int errors = 0;
    int failures = 0;
    int disabled = 0;
    int skipped = 0;
    for (TestData result : tests) {
      if (result.isError()) {
        errors++;
      }
      if (result.isFailure()) {
        failures++;
      }
      if (result.isDisabled()) {
        disabled++;
      }
      if (result.isSkipped()) {
        skipped++;
      }
    }
    xml.writeAttribute("failures", String.valueOf(failures));
    xml.writeAttribute("errors", String.valueOf(errors));
    xml.writeAttribute("disabled", String.valueOf(disabled));
    xml.writeAttribute("skipped", String.valueOf(skipped));

    // The bazel junit4 test runner seems to leave these values empty.
    // Emulating that somewhat strange behaviour here.
    xml.writeAttribute("package", "");
    xml.writeEmptyElement("properties");

    for (TestData testCase : tests) {
      testRenderer.toXml(xml, testCase);
    }

    writeTextElement(xml, "system-out", suite.getStdOut());
    writeTextElement(xml, "system-err", suite.getStdErr());

    xml.writeEndElement();
  }
}

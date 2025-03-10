package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static com.github.bazel_contrib.contrib_rules_jvm.junit5.SafeXml.escapeIllegalCharacters;
import static com.github.bazel_contrib.contrib_rules_jvm.junit5.SafeXml.writeTextElement;

import java.net.InetAddress;
import java.time.format.DateTimeFormatter;
import java.text.DecimalFormat;
import java.text.DecimalFormatSymbols;
import java.time.Duration;
import java.time.Instant;
import java.util.Collection;
import java.util.Locale;
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

    xml.writeAttribute("name", escapeIllegalCharacters(suite.getId().getLegacyReportingName()));
    xml.writeAttribute("timestamp", DateTimeFormatter.ISO_INSTANT.format(suite.getStarted()));
    xml.writeAttribute("hostname", getHostname());
    xml.writeAttribute("tests", String.valueOf(tests.size()));

    DecimalFormat decimalFormat = new DecimalFormat("#.##", new DecimalFormatSymbols(Locale.ROOT));
    /* @Nullable */ Duration maybeDuration = suite.getDuration();
    Duration duration =
            maybeDuration == null ? Duration.between(suite.getStarted(), Instant.now()) : maybeDuration;
    xml.writeAttribute("time", decimalFormat.format(duration.toMillis() / 1000f));

    int errors = 0;
    int failures = 0;
    int disabled = 0;
    int skipped = 0;
    for (TestData result : tests) {
      // Tests which didn't complete are considered to be failures.
      // The caller is expected to filter out still-running tests, so if we got here,
      // it's because the test has been cancelled (e.g. because of a timeout).
      if (result.getDuration() == null) {
        failures++;
      } else {
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

  private String getHostname() {
    try {
      return InetAddress.getLocalHost().getHostName();
    } catch (Exception e) {
      return "localhost";
    }
  }
}

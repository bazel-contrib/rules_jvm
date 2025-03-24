package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static com.github.bazel_contrib.contrib_rules_jvm.junit5.SafeXml.escapeIllegalCharacters;
import static com.github.bazel_contrib.contrib_rules_jvm.junit5.SafeXml.writeTextElement;

import java.io.PrintWriter;
import java.io.StringWriter;
import java.math.RoundingMode;
import java.text.DecimalFormat;
import java.text.DecimalFormatSymbols;
import java.time.Duration;
import java.time.Instant;
import java.util.Locale;
import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamWriter;
import org.junit.platform.engine.TestExecutionResult;
import org.junit.platform.launcher.TestIdentifier;
import org.junit.platform.launcher.TestPlan;
import org.junit.platform.reporting.legacy.LegacyReportingUtils;

class TestCaseXmlRenderer {

  private static final DecimalFormatSymbols DECIMAL_FORMAT_SYMBOLS =
      new DecimalFormatSymbols(Locale.ROOT);
  private final TestPlan testPlan;

  public TestCaseXmlRenderer(TestPlan testPlan) {
    this.testPlan = testPlan;
  }

  public void toXml(XMLStreamWriter xml, TestData test) throws XMLStreamException {
    DecimalFormat decimalFormat = new DecimalFormat("#.##", DECIMAL_FORMAT_SYMBOLS);
    decimalFormat.setRoundingMode(RoundingMode.HALF_UP);

    TestIdentifier id = test.getId();

    String name = getTestName(test);

    xml.writeStartElement("testcase");
    xml.writeAttribute("name", escapeIllegalCharacters(name));
    xml.writeAttribute("classname", LegacyReportingUtils.getClassName(testPlan, id));

    /* @Nullable */ Duration maybeDuration = test.getDuration();
    boolean wasInterrupted = maybeDuration == null;
    Duration duration =
        maybeDuration == null ? Duration.between(test.getStarted(), Instant.now()) : maybeDuration;
    xml.writeAttribute("time", decimalFormat.format(duration.toMillis() / 1000f));

    if (wasInterrupted) {
      xml.writeStartElement("failure");
      xml.writeCData("Test timed out and was interrupted");
      xml.writeEndElement();
    } else {
      if (test.isDisabled() || test.isSkipped()) {
        xml.writeStartElement("skipped");
        if (test.getSkipReason() != null) {
          xml.writeCData(test.getSkipReason());
        } else {
          writeThrowableMessage(xml, test.getResult());
        }
        xml.writeEndElement();
      }
      if (test.isFailure() || test.isError()) {
        xml.writeStartElement(test.isFailure() ? "failure" : "error");
        writeThrowableMessage(xml, test.getResult());
        xml.writeEndElement();
      }
    }

    writeTextElement(xml, "system-out", test.getStdOut());
    writeTextElement(xml, "system-err", test.getStdErr());

    xml.writeEndElement();
  }

  private void writeThrowableMessage(XMLStreamWriter xml, TestExecutionResult result)
      throws XMLStreamException {
    Throwable throwable = null;
    if (result != null) {
      throwable = result.getThrowable().orElse(null);
    }
    if (throwable == null) {
      // Stub out the values
      xml.writeAttribute("message", "unknown cause");
      xml.writeAttribute("type", RuntimeException.class.getName());
      return;
    }

    xml.writeAttribute("message", escapeIllegalCharacters(String.valueOf(throwable.getMessage())));
    xml.writeAttribute("type", throwable.getClass().getName());

    StringWriter stringWriter = new StringWriter();
    throwable.printStackTrace(new PrintWriter(stringWriter));

    xml.writeCData(escapeIllegalCharacters(stringWriter.toString()));
  }

  private String getTestName(TestData test) {
    TestIdentifier id = test.getId();
    // checking for the '[' as a proxy for the ordinal parameterized test.
    // If there is some edge case not considered, here, it should still be okay, as it really just
    // formats the test case name.
    if (id.getParentId().isPresent() && id.getDisplayName().startsWith("[")) {
      return getCustomDisplayName(id);
    }
    // this check assumes ParameterizedTest is dynamic, leaving this for now but may not need it.
    // Edge case maybe when there is no parent (if ever possible)
    if (test.isDynamic()) {
      return id.getDisplayName(); // [ordinal] name=value...
    } else {
      String name = id.getLegacyReportingName();
      int index = name.indexOf('(');
      if (index != -1) {
        name = name.substring(0, index);
      }
      return name;
    }
  }

  private String getCustomDisplayName(TestIdentifier testIdentifier) {
    TestIdentifier parent = testPlan.getTestIdentifier(testIdentifier.getParentId().get());
    String methodName = parent.getDisplayName().replaceAll("\\(\\)", "");
    return methodName + " " + testIdentifier.getDisplayName();
  }
}

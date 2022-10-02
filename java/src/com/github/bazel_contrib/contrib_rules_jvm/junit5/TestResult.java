package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.io.PrintWriter;
import java.io.StringWriter;
import java.math.RoundingMode;
import java.text.DecimalFormat;
import java.util.Optional;
import javax.xml.stream.XMLStreamWriter;
import org.junit.AssumptionViolatedException;
import org.junit.platform.engine.TestExecutionResult;
import org.junit.platform.launcher.TestIdentifier;
import org.junit.platform.launcher.TestPlan;
import org.junit.platform.reporting.legacy.LegacyReportingUtils;
import org.opentest4j.TestAbortedException;

class TestResult extends BaseResult {
  private final TestPlan testPlan;
  private final boolean isDynamic;

  public TestResult(TestPlan testPlan, TestIdentifier id, boolean isDynamic) {
    super(id);
    this.testPlan = testPlan;
    this.isDynamic = isDynamic;
  }

  public boolean isError() {
    TestExecutionResult result = getResult();
    if (result == null || result.getStatus() == TestExecutionResult.Status.SUCCESSFUL) {
      return false;
    }

    return result.getThrowable().map(thr -> thr instanceof AssertionError).orElse(false);
  }

  public boolean isFailure() {
    TestExecutionResult result = getResult();
    if (result == null || result.getStatus() == TestExecutionResult.Status.SUCCESSFUL) {
      return false;
    }

    return result.getThrowable().map(thr -> (!(thr instanceof AssertionError))).orElse(false);
  }

  public boolean isSkipped() {
    return getResult()
        .getThrowable()
        .map(
            thr ->
                (thr instanceof TestAbortedException || thr instanceof AssumptionViolatedException))
        .orElse(false);
  }

  @Override
  public void toXml(XMLStreamWriter xml) {
    DecimalFormat decimalFormat = new DecimalFormat("#.##");
    decimalFormat.setRoundingMode(RoundingMode.HALF_UP);

    write(
        () -> {
          String name;
          if (isDynamic) {
            name = getTestId().getDisplayName(); // [ordinal] name=value...
          } else {
            // Massage the name
            name = getTestId().getLegacyReportingName();
            int index = name.indexOf('(');
            if (index != -1) {
              name = name.substring(0, index);
            }
          }

          xml.writeStartElement("testcase");
          xml.writeAttribute("name", name);
          xml.writeAttribute("classname", LegacyReportingUtils.getClassName(testPlan, getTestId()));
          xml.writeAttribute("time", decimalFormat.format(getDuration().toMillis() / 1000f));

          if (isSkipped()) {
            xml.writeStartElement("skipped");
          }
          if (isFailure() || isError()) {
            Throwable throwable = getResult().getThrowable().orElse(null);

            xml.writeStartElement(isFailure() ? "failure" : "error");
            if (throwable == null) {
              // Stub out the values
              xml.writeAttribute("message", "unknown cause");
              xml.writeAttribute("type", RuntimeException.class.getName());
              xml.writeEndElement();
              return;
            }

            xml.writeAttribute("message", throwable.getMessage());
            xml.writeAttribute("type", throwable.getClass().getName());

            StringWriter stringWriter = new StringWriter();
            throwable.printStackTrace(new PrintWriter(stringWriter));

            xml.writeCData(stringWriter.toString());

            xml.writeEndElement();
          }

          String stdout = getStdOut();
          if (stdout != null) {
            xml.writeStartElement("system-out");
            xml.writeCData(stdout);
            xml.writeEndElement();
          }

          String stderr = getStdErr();
          if (stderr != null) {
            xml.writeStartElement("system-err");
            xml.writeCData(stderr);
            xml.writeEndElement();
          }

          xml.writeEndElement();
        });
  }

  @Override
  protected Optional<BaseResult> get(TestIdentifier id) {
    return getTestId().equals(id) ? Optional.of(this) : Optional.empty();
  }
}

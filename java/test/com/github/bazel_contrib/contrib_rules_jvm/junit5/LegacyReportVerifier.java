package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.io.File;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.w3c.dom.NodeList;

public class LegacyReportVerifier {

  public static void main(String[] args) throws Exception {
    if (args.length < 1) {
      System.err.println("Usage: LegacyReportVerifier <report-file>");
      System.exit(1);
    }
    File reportFile = new File(args[0]);
    if (!reportFile.exists()) {
      System.err.println("FAIL: Report file does not exist: " + args[0]);
      System.exit(1);
    }

    DocumentBuilder builder = DocumentBuilderFactory.newInstance().newDocumentBuilder();
    Document doc = builder.parse(reportFile);

    NodeList testsuites = doc.getElementsByTagName("testsuite");
    if (testsuites.getLength() == 0) {
      System.err.println("FAIL: No <testsuite> element found");
      System.exit(1);
    }

    Element testsuite = (Element) testsuites.item(0);
    checkAttribute(testsuite, "tests", "1");
    checkAttribute(testsuite, "failures", "0");
    checkAttribute(testsuite, "errors", "0");
    checkAttribute(testsuite, "skipped", "0");

    NodeList testcases = doc.getElementsByTagName("testcase");
    if (testcases.getLength() == 0) {
      System.err.println("FAIL: No <testcase> elements found");
      System.exit(1);
    }

    Element testcase = (Element) testcases.item(0);
    String testcaseName = testcase.getAttribute("name");
    if (!testcaseName.contains("shouldUseLegacyReportGenerator")) {
      System.err.println(
          "FAIL: <testcase @name> does not contain 'shouldUseLegacyReportGenerator': "
              + testcaseName);
      System.exit(1);
    }

    String testcaseClassname = testcase.getAttribute("classname");
    if (!testcaseClassname.contains("LegacyReportGeneratorTest")) {
      System.err.println(
          "FAIL: <testcase @classname> does not contain 'LegacyReportGeneratorTest': "
              + testcaseClassname);
      System.exit(1);
    }

    System.out.println("PASS: Legacy XML report structure is correct");
  }

  private static void checkAttribute(Element element, String attr, String expected) {
    String actual = element.getAttribute(attr);
    if (!expected.equals(actual)) {
      System.err.println(
          "FAIL: <testsuite @" + attr + "> expected '" + expected + "' but was '" + actual + "'");
      System.exit(1);
    }
  }
}

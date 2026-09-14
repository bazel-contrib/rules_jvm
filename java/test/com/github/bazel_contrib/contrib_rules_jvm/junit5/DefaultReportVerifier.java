package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.io.File;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.w3c.dom.NodeList;

public class DefaultReportVerifier {

  public static void main(String[] args) throws Exception {
    if (args.length < 1) {
      System.err.println("Usage: DefaultReportVerifier <xml-output-file>");
      System.exit(1);
    }
    File xmlFile = new File(args[0]);
    if (!xmlFile.exists() || xmlFile.length() == 0) {
      System.err.println("FAIL: XML output file is missing or empty: " + args[0]);
      System.exit(1);
    }

    DocumentBuilder builder = DocumentBuilderFactory.newInstance().newDocumentBuilder();
    Document doc = builder.parse(xmlFile);

    // Standard Bazel output wraps suites in <testsuites>; find the <testsuite> for our class.
    NodeList testsuites = doc.getElementsByTagName("testsuite");
    if (testsuites.getLength() == 0) {
      System.err.println("FAIL: No <testsuite> element found in XML output");
      System.exit(1);
    }

    Element testsuite = null;
    for (int i = 0; i < testsuites.getLength(); i++) {
      Element candidate = (Element) testsuites.item(i);
      if (candidate.getAttribute("name").contains("DefaultReportGeneratorTest")) {
        testsuite = candidate;
        break;
      }
    }
    if (testsuite == null) {
      System.err.println(
          "FAIL: No <testsuite> found with name containing 'DefaultReportGeneratorTest'");
      System.exit(1);
    }

    checkAttribute(testsuite, "tests", "1");
    checkAttribute(testsuite, "failures", "0");
    checkAttribute(testsuite, "errors", "0");
    checkAttribute(testsuite, "skipped", "0");

    NodeList testcases = testsuite.getElementsByTagName("testcase");
    if (testcases.getLength() == 0) {
      System.err.println("FAIL: No <testcase> elements found");
      System.exit(1);
    }

    Element testcase = (Element) testcases.item(0);
    String testcaseName = testcase.getAttribute("name");
    if (!testcaseName.contains("shouldUseDefaultReportGenerator")) {
      System.err.println(
          "FAIL: <testcase @name> does not contain 'shouldUseDefaultReportGenerator': "
              + testcaseName);
      System.exit(1);
    }

    String testcaseClassname = testcase.getAttribute("classname");
    if (!testcaseClassname.contains("DefaultReportGeneratorTest")) {
      System.err.println(
          "FAIL: <testcase @classname> does not contain 'DefaultReportGeneratorTest': "
              + testcaseClassname);
      System.exit(1);
    }

    System.out.println("PASS: Default XML report structure is correct");
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

package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.io.File;
import javax.xml.parsers.DocumentBuilder;
import javax.xml.parsers.DocumentBuilderFactory;
import org.w3c.dom.Document;
import org.w3c.dom.Element;
import org.w3c.dom.NodeList;

/**
 * Verifies an Open Test Reporting XML file produced by OpenTestReportGeneratingListener.
 *
 * <p>Uses non-namespace-aware DOM so that qualified element names (e:started, e:finished,
 * java:methodSource) are matched literally, avoiding dependency on specific namespace URI versions.
 */
public class OpenReportVerifier {

  public static void main(String[] args) throws Exception {
    if (args.length < 1) {
      System.err.println("Usage: OpenReportVerifier <report-file>");
      System.exit(1);
    }
    File reportFile = new File(args[0]);
    if (!reportFile.exists()) {
      System.err.println("FAIL: Report file does not exist: " + args[0]);
      System.exit(1);
    }

    DocumentBuilder builder = DocumentBuilderFactory.newInstance().newDocumentBuilder();
    Document doc = builder.parse(reportFile);

    NodeList startedElements = doc.getElementsByTagName("e:started");

    String testEventId = null;
    for (int i = 0; i < startedElements.getLength(); i++) {
      Element started = (Element) startedElements.item(i);
      if (started.getAttribute("name").contains("shouldUseOpenReportGenerator")) {
        testEventId = started.getAttribute("id");
        NodeList methodSources = started.getElementsByTagName("java:methodSource");
        if (methodSources.getLength() > 0) {
          String className = ((Element) methodSources.item(0)).getAttribute("className");
          if (!className.contains("OpenReportGeneratorTest")) {
            System.err.println(
                "FAIL: java:methodSource/@className does not contain 'OpenReportGeneratorTest': "
                    + className);
            System.exit(1);
          }
        }
        break;
      }
    }

    if (testEventId == null) {
      System.err.println(
          "FAIL: No e:started element found with name containing 'shouldUseOpenReportGenerator'");
      System.exit(1);
    }

    NodeList finishedElements = doc.getElementsByTagName("e:finished");
    boolean foundSuccessful = false;
    for (int i = 0; i < finishedElements.getLength(); i++) {
      Element finished = (Element) finishedElements.item(i);
      if (testEventId.equals(finished.getAttribute("id"))) {
        NodeList results = finished.getElementsByTagName("result");
        if (results.getLength() > 0) {
          String status = ((Element) results.item(0)).getAttribute("status");
          if (!"SUCCESSFUL".equals(status)) {
            System.err.println("FAIL: Test result status is not SUCCESSFUL: " + status);
            System.exit(1);
          }
          foundSuccessful = true;
        }
        break;
      }
    }

    if (!foundSuccessful) {
      System.err.println(
          "FAIL: No successful e:finished element found for test id=" + testEventId);
      System.exit(1);
    }

    System.out.println("PASS: Open test report structure is correct");
  }
}

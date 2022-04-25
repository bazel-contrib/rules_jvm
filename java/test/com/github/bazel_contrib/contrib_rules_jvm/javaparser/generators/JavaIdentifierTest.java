package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static org.junit.jupiter.api.Assertions.assertEquals;

import org.junit.jupiter.api.Test;

public class JavaIdentifierTest {

  @Test
  void testToString() {
    JavaIdentifier id =
        new JavaIdentifier(
            "com.gazelle.java.javaparser.generators",
            "JavaIdentifier",
            "artifact(\"com.gazelle.java.javaparser:generators\")");
    String testString =
        "com.gazelle.java.javaparser.generators.JavaIdentifier ->"
            + " artifact(\"com.gazelle.java.javaparser:generators\")";
    assertEquals(id.toString(), testString);
  }

  @Test
  void testHashCode() {
    JavaIdentifier id =
        new JavaIdentifier(
            "com.gazelle.java.javaparser.generators",
            "JavaIdentifier",
            "artifact(\"com.gazelle.java.javaparser:generators\")");
    int hashCode = 1986351184;
    assertEquals(id.hashCode(), hashCode);
  }
}

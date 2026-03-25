package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators.ClassNames.isLikelyClassName;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.Test;

public class ClassNamesTest {

  @Test
  void pascalCaseNamesAreClassNames() {
    assertTrue(isLikelyClassName("String"));
    assertTrue(isLikelyClassName("ArrayList"));
    assertTrue(isLikelyClassName("KtParser"));
    assertTrue(isLikelyClassName("IOException"));
  }

  @Test
  void emptyStringIsNotClassName() {
    assertFalse(isLikelyClassName(""));
  }

  @Test
  void lowercaseStartIsNotClassName() {
    assertFalse(isLikelyClassName("string"));
    assertFalse(isLikelyClassName("myClass"));
    assertFalse(isLikelyClassName("parseInt"));
  }

  @Test
  void singleUppercaseCharIsNotClassName() {
    // Single uppercase letters are typically type parameters (T, E, K, V)
    assertFalse(isLikelyClassName("T"));
    assertFalse(isLikelyClassName("E"));
  }

  @Test
  void allCapsConstantsAreNotClassNames() {
    assertFalse(isLikelyClassName("MAX_VALUE"));
    assertFalse(isLikelyClassName("HTTP"));
    assertFalse(isLikelyClassName("SQL"));
    assertFalse(isLikelyClassName("IO"));
  }

  @Test
  void twoCharPascalCaseIsClassName() {
    assertTrue(isLikelyClassName("Io"));
    assertTrue(isLikelyClassName("Ok"));
  }
}

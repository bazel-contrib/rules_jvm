package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.util.function.Predicate;
import java.util.regex.Pattern;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import org.junit.platform.engine.FilterResult;
import org.junit.platform.engine.TestDescriptor;
import org.junit.platform.engine.TestSource;
import org.junit.platform.engine.support.descriptor.ClassSource;
import org.junit.platform.engine.support.descriptor.MethodSource;
import org.junit.platform.launcher.PostDiscoveryFilter;

/**
 * Attempts to mirror the logic from Bazel's own
 * com.google.testing.junit.junit4.runner.RegExTestCaseFilter, which forms a string of the test
 * class name and the method name.
 */
public class PatternFilter implements PostDiscoveryFilter {

  private final String rawPattern;
  private final Predicate<String> pattern;

  public PatternFilter(String pattern) {
    if (pattern == null || pattern.isEmpty()) {
      pattern = ".*";
    } else {
      pattern = convertCommaSeparatedSelections(pattern);
    }

    this.rawPattern = pattern;
    this.pattern = Pattern.compile(pattern).asPredicate();
  }

  @Override
  public FilterResult apply(TestDescriptor object) {
    if (!object.isTest()) {
      return FilterResult.included("Including container: " + object.getDisplayName());
    }

    if (object.getSource().isEmpty()) {
      return FilterResult.excluded("Skipping a test without a source: " + object.getDisplayName());
    }

    TestSource source = object.getSource().get();
    String testName;
    if (source instanceof MethodSource) {
      MethodSource method = (MethodSource) source;
      testName = method.getClassName() + "#" + method.getMethodName();
    } else if (source instanceof ClassSource) {
      ClassSource clazz = (ClassSource) source;
      testName = clazz.getClassName() + "#";
    } else {
      testName = object.getDisplayName();
    }

    if (pattern.test(testName)) {
      return FilterResult.included("Matched " + testName + " by " + rawPattern);
    }

    return FilterResult.excluded("Did not match " + rawPattern);
  }

  /**
   * Converts comma-separated selections in patterns like:
   *
   * <ul>
   *   <li>classes: "path.to.SomeTest,path.to.AnotherTest" -> "path.to.SomeTest|path.to.AnotherTest"
   *   <li>methods: "path.to.SomeTest#testSomething,testSomethingElse" ->
   *       "path.to.SomeTest#testSomething$|path.to.SomeTest#testSomethingElse$"
   * </ul>
   */
  private static String convertCommaSeparatedSelections(String pattern) {
    var selections = pattern.split(",");
    if (selections.length == 1) {
      return ensureExactMethodName(pattern);
    }
    var precedingClassSelection = selections[0];
    var precedingHashIndex = precedingClassSelection.indexOf('#');
    for (int i = 1; i < selections.length; i++) {
      var selection = selections[i];
      var hashIndex = selection.indexOf('#');
      if (hashIndex > -1) { // `class#` or `class#method`
        precedingClassSelection = selection;
        precedingHashIndex = hashIndex;
      } else if (precedingHashIndex > -1) { // prepend preceding `class#`
        selections[i] = precedingClassSelection.substring(0, precedingHashIndex + 1) + selection;
      }
    }
    return Stream.of(selections)
        .map(PatternFilter::ensureExactMethodName)
        .collect(Collectors.joining("|"));
  }

  /** Appends '$' to patterns like "class#method" or "#method", unless already done. */
  private static String ensureExactMethodName(String pattern) {
    return pattern.matches(".*#.*[^$]$") ? pattern + '$' : pattern;
  }
}

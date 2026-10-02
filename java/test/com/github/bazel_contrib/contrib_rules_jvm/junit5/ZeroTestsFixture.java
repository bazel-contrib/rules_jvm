package com.github.bazel_contrib.contrib_rules_jvm.junit5;

/**
 * A fixture class with no {@code @Test} methods, used by {@link EmptyTestResultsTest} to verify
 * that {@link ActualRunner} fails when a test class matches zero tests. Deliberately does not
 * end in "Test.java" so it is not itself picked up as a test by {@code java_test_suite}.
 */
public class ZeroTestsFixture {
  public void notATest() {}
}

package com.github.bazel_contrib.contrib_rules_jvm.comparative;

import java.util.stream.Stream;
import org.junit.jupiter.api.DynamicTest;
import org.junit.jupiter.api.TestFactory;
import org.junit.jupiter.api.function.Executable;

public class DynamicTests {

  @TestFactory
  Stream<DynamicTest> translateDynamicTestsFromStream() {
    return Stream.of("one", "two", "three")
        .map(
            word ->
                DynamicTest.dynamicTest(
                    "Test of " + word,
                    // Use an anonymous class here to prevent spotbugs getting annoyed
                    new Executable() {
                      @Override
                      public void execute() {}
                    }));
  }
}

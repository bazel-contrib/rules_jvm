package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.lang.reflect.Method;
import java.util.Optional;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.engine.config.DefaultJupiterConfiguration;
import org.junit.jupiter.engine.config.JupiterConfiguration;
import org.junit.jupiter.engine.descriptor.ClassTestDescriptor;
import org.junit.jupiter.engine.descriptor.JupiterEngineDescriptor;
import org.junit.jupiter.engine.descriptor.TestMethodTestDescriptor;
import org.junit.jupiter.engine.descriptor.TestTemplateTestDescriptor;
import org.junit.platform.engine.ConfigurationParameters;
import org.junit.platform.engine.FilterResult;
import org.junit.platform.engine.UniqueId;

public class FilteringTest {

  private JupiterEngineDescriptor engineDescriptor;
  private ClassTestDescriptor classTestDescriptor;
  private ClassTestDescriptor nestedClassTestDescriptor;
  private TestMethodTestDescriptor testMethodTestDescriptor;
  private TestTemplateTestDescriptor testTemplateTestDescriptor;
  private TestMethodTestDescriptor siblingTestMethodTestDescriptor;
  private TestMethodTestDescriptor nestedTestMethodTestDescriptor;

  @BeforeEach
  public void setup() throws NoSuchMethodException {
    UniqueId engine = UniqueId.forEngine("engine");
    JupiterConfiguration config = new DefaultJupiterConfiguration(new EmptyConfigParameters());

    engineDescriptor = new JupiterEngineDescriptor(engine, config);
    UniqueId classId = engine.append("class", "foo");
    classTestDescriptor = new ClassTestDescriptor(classId, JUnit5StyleTest.class, config);
    Method method = JUnit5StyleTest.class.getMethod("alwaysPasses");
    testMethodTestDescriptor =
        new TestMethodTestDescriptor(
            classId.append("method", "bar"), JUnit5StyleTest.class, method, config);
    testTemplateTestDescriptor =
        new TestTemplateTestDescriptor(
            classId.append("method", "baz"), JUnit5StyleTest.class, method, config);
    Method siblingMethod = JUnit5StyleTest.class.getMethod("alwaysPassesToo");
    siblingTestMethodTestDescriptor =
        new TestMethodTestDescriptor(
            classId.append("method", "qux"), JUnit5StyleTest.class, siblingMethod, config);
    nestedClassTestDescriptor =
        new ClassTestDescriptor(classId, JUnit5StyleTest.NestedTest.class, config);
    Method nestedMethod = JUnit5StyleTest.NestedTest.class.getMethod("alwaysPassesToo");
    nestedTestMethodTestDescriptor =
        new TestMethodTestDescriptor(
            classId.append("method", "quux"),
            JUnit5StyleTest.NestedTest.class,
            nestedMethod,
            config);
  }

  @Test
  public void ifFilterIsNotSetAllTestsShouldBeAccepted() {
    PatternFilter filter = new PatternFilter(null);

    FilterResult engineResult = filter.apply(engineDescriptor);
    assertTrue(engineResult.included());

    FilterResult classResult = filter.apply(classTestDescriptor);
    assertTrue(classResult.included());

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertTrue(testResult.included());

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertTrue(dynamicTestResult.included());

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertTrue(siblingTestResult.included());

    FilterResult nestedClassResult = filter.apply(nestedClassTestDescriptor);
    assertTrue(nestedClassResult.included());

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertTrue(nestedTestResult.included());
  }

  @Test
  public void ifFilterIsSetButEmptyAllTestsShouldBeAccepted() {
    PatternFilter filter = new PatternFilter("");

    FilterResult engineResult = filter.apply(engineDescriptor);
    assertTrue(engineResult.included());

    FilterResult classResult = filter.apply(classTestDescriptor);
    assertTrue(classResult.included());

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertTrue(testResult.included());

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertTrue(dynamicTestResult.included());

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertTrue(siblingTestResult.included());

    FilterResult nestedClassResult = filter.apply(nestedClassTestDescriptor);
    assertTrue(nestedClassResult.included());

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertTrue(nestedTestResult.included());
  }

  @Test
  public void ifFilterIsSetButNoTestsMatchTheContainersAreIncluded() {
    PatternFilter filter = new PatternFilter("com.example.will.never.Match#");

    FilterResult engineResult = filter.apply(engineDescriptor);
    assertTrue(engineResult.included());

    FilterResult classResult = filter.apply(classTestDescriptor);
    assertTrue(classResult.included());

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertFalse(testResult.included());

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertFalse(dynamicTestResult.included());

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertFalse(siblingTestResult.included());

    FilterResult nestedClassResult = filter.apply(nestedClassTestDescriptor);
    assertTrue(nestedClassResult.included());

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertFalse(nestedTestResult.included());
  }

  @Test
  public void shouldIncludeATestMethodIfTheFilterIsJustTheClassName() {
    PatternFilter filter = new PatternFilter(JUnit5StyleTest.class.getName() + "#");

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertTrue(testResult.included());

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertTrue(dynamicTestResult.included());

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertTrue(siblingTestResult.included());

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertFalse(nestedTestResult.included(), "method in nested class should not be matched");
  }

  @Test
  public void shouldIncludeANestedTestMethodIfTheFilterIsJustTheNestedClassName() {
    PatternFilter filter = new PatternFilter(JUnit5StyleTest.NestedTest.class.getName() + "#");

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertFalse(testResult.included(), "method in enclosing class should not be matched");

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertFalse(dynamicTestResult.included(), "method in enclosing class should not be matched");

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertFalse(siblingTestResult.included(), "method in enclosing class should not be matched");

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertTrue(nestedTestResult.included());
  }

  @Test
  public void shouldIncludeMultipleTestMethodsIfTheFilterComprisesMultipleClassNames() {
    PatternFilter filter =
        new PatternFilter(
            String.join(
                "|", JUnit5StyleTest.class.getName(), JUnit5StyleTest.NestedTest.class.getName()));

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertTrue(testResult.included());

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertTrue(dynamicTestResult.included());

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertTrue(siblingTestResult.included());

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertTrue(nestedTestResult.included());
  }

  @Test
  public void shouldIncludeMultipleTestMethodsIfTheFilterComprisesMultipleClassNamesWithCommas() {
    PatternFilter filter =
        new PatternFilter(
            String.join(
                ",", JUnit5StyleTest.class.getName(), JUnit5StyleTest.NestedTest.class.getName()));

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertTrue(testResult.included());

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertTrue(dynamicTestResult.included());

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertTrue(siblingTestResult.included());

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertTrue(nestedTestResult.included());
  }

  @Test
  public void shouldNotIncludeATestMethodIfTheFilterDoesNotMatchTheMethodName() {
    PatternFilter filter = new PatternFilter("#foo");

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertFalse(testResult.included());

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertFalse(dynamicTestResult.included());

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertFalse(siblingTestResult.included());

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertFalse(nestedTestResult.included());
  }

  @Test
  public void shouldIncludeATestMethodIfTheFilterMatchesTheExactShortMethodName() {
    PatternFilter filter = new PatternFilter("#alwaysPasses");

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertTrue(testResult.included());

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertTrue(dynamicTestResult.included());

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertFalse(siblingTestResult.included(), "longer method name should not be matched");

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertFalse(nestedTestResult.included(), "longer method name should not be matched");
  }

  @Test
  public void shouldIncludeATestMethodIfTheFilterMatchesTheExactLongMethodName() {
    PatternFilter filter = new PatternFilter("#alwaysPassesToo");

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertFalse(testResult.included(), "shorter method name should not be matched");

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertFalse(dynamicTestResult.included(), "shorter method name should not be matched");

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertTrue(siblingTestResult.included());

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertTrue(nestedTestResult.included());
  }

  @Test
  public void shouldIncludeMultipleTestMethodsIfTheFilterComprisesMultipleMethodNames() {
    PatternFilter filter = new PatternFilter("JUnit5StyleTest#alwaysPasses,alwaysPassesToo");

    FilterResult testResult = filter.apply(testMethodTestDescriptor);
    assertTrue(testResult.included());

    FilterResult dynamicTestResult = filter.apply(testTemplateTestDescriptor);
    assertTrue(dynamicTestResult.included());

    FilterResult siblingTestResult = filter.apply(siblingTestMethodTestDescriptor);
    assertTrue(siblingTestResult.included());

    FilterResult nestedTestResult = filter.apply(nestedTestMethodTestDescriptor);
    assertFalse(nestedTestResult.included(), "nested class should not be matched");
  }

  private static class EmptyConfigParameters implements ConfigurationParameters {
    @Override
    public Optional<String> get(String key) {
      return Optional.empty();
    }

    @Override
    public Optional<Boolean> getBoolean(String key) {
      return Optional.empty();
    }

    @Override
    public int size() {
      return 0;
    }
  }

  private static class JUnit5StyleTest {
    @Test
    public void alwaysPasses() {}

    @Test
    public void alwaysPassesToo() {}

    @Nested
    private class NestedTest {
      @Test
      public void alwaysPassesToo() {}
    }
  }
}

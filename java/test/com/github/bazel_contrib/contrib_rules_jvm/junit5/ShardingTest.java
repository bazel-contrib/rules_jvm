package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;

import com.github.bazel_contrib.contrib_rules_jvm.junit5.sample.ShardingTestMoreTests;
import com.github.bazel_contrib.contrib_rules_jvm.junit5.sample.ShardingTestTests;
import java.util.HashSet;
import java.util.Set;
import org.junit.Before;
import org.junit.Test;
import org.junit.platform.engine.UniqueId;
import org.junit.platform.engine.discovery.DiscoverySelectors;
import org.junit.platform.launcher.LauncherDiscoveryRequest;
import org.junit.platform.launcher.PostDiscoveryFilter;
import org.junit.platform.launcher.TestExecutionListener;
import org.junit.platform.launcher.TestIdentifier;
import org.junit.platform.launcher.core.LauncherConfig;
import org.junit.platform.launcher.core.LauncherDiscoveryRequestBuilder;
import org.junit.platform.launcher.core.LauncherFactory;

public class ShardingTest {

  private static final UniqueId ENGINE_ID = UniqueId.forEngine("junit-jupiter");
  private static final UniqueId SHARDING_TEST_TESTS =
      ENGINE_ID.append("class", ShardingTestTests.class.getName());
  private static final UniqueId SHARDING_TEST_MORE_TESTS =
      ENGINE_ID.append("class", ShardingTestMoreTests.class.getName());

  private Set<UniqueId> tests;

  @Before
  public void setUp() {
    LauncherDiscoveryRequest request =
        LauncherDiscoveryRequestBuilder.request()
            .selectors(
                DiscoverySelectors.selectPackage(
                    ShardingTest.class.getPackage().getName() + ".sample"))
            .build();

    tests = new HashSet<>();
    PostDiscoveryFilter shardFilter = TestSharding.makeShardFilter();
    LauncherConfig config =
        LauncherConfig.builder()
            .addTestExecutionListeners(
                new TestExecutionListener() {
                  @Override
                  public void executionStarted(TestIdentifier testIdentifier) {
                    tests.add(testIdentifier.getUniqueIdObject());
                  }
                })
            .addPostDiscoveryFilters(shardFilter)
            .build();

    LauncherFactory.create(config).execute(request);
  }

  @Test
  public void testShard1() {
    assertEquals(
        Set.of(
            ENGINE_ID,
            SHARDING_TEST_TESTS,
            test(SHARDING_TEST_TESTS, "testSeven"),
            SHARDING_TEST_MORE_TESTS,
            test(SHARDING_TEST_MORE_TESTS, "testTwo"),
            testTemplate(SHARDING_TEST_MORE_TESTS, "testParameterized", "int"),
            testTemplateInvocation(SHARDING_TEST_MORE_TESTS, "testParameterized", "int", 1),
            testTemplateInvocation(SHARDING_TEST_MORE_TESTS, "testParameterized", "int", 2),
            testTemplateInvocation(SHARDING_TEST_MORE_TESTS, "testParameterized", "int", 3),
            testTemplateInvocation(SHARDING_TEST_MORE_TESTS, "testParameterized", "int", 4),
            testTemplateInvocation(SHARDING_TEST_MORE_TESTS, "testParameterized", "int", 5),
            testTemplateInvocation(SHARDING_TEST_MORE_TESTS, "testParameterized", "int", 6),
            testTemplateInvocation(SHARDING_TEST_MORE_TESTS, "testParameterized", "int", 7),
            testTemplateInvocation(SHARDING_TEST_MORE_TESTS, "testParameterized", "int", 8)),
        tests);
  }

  @Test
  public void testShard2() {
    assertEquals(
        Set.of(
            ENGINE_ID,
            SHARDING_TEST_TESTS,
            test(SHARDING_TEST_TESTS, "testOne"),
            test(SHARDING_TEST_TESTS, "testThree"),
            test(SHARDING_TEST_TESTS, "testFour"),
            test(SHARDING_TEST_TESTS, "testFive"),
            test(SHARDING_TEST_TESTS, "testSix"),
            SHARDING_TEST_MORE_TESTS,
            test(SHARDING_TEST_MORE_TESTS, "testSeven")),
        tests);
  }

  @Test
  public void testShard3() {
    assertEquals(
        Set.of(
            ENGINE_ID,
            SHARDING_TEST_TESTS,
            test(SHARDING_TEST_TESTS, "testEight"),
            SHARDING_TEST_MORE_TESTS,
            test(SHARDING_TEST_MORE_TESTS, "testOne"),
            test(SHARDING_TEST_MORE_TESTS, "testThree"),
            test(SHARDING_TEST_MORE_TESTS, "testFour"),
            test(SHARDING_TEST_MORE_TESTS, "testFive"),
            test(SHARDING_TEST_MORE_TESTS, "testSix")),
        tests);
  }

  @Test
  public void testShard4() {
    assertEquals(
        Set.of(
            ENGINE_ID,
            SHARDING_TEST_TESTS,
            test(SHARDING_TEST_TESTS, "testTwo"),
            testTemplate(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo"),
            testTemplateInvocation(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo", 1),
            testTemplateInvocation(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo", 2),
            testTemplateInvocation(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo", 3),
            testTemplateInvocation(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo", 4),
            testTemplateInvocation(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo", 5),
            testTemplateInvocation(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo", 6),
            testTemplateInvocation(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo", 7),
            testTemplateInvocation(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo", 8),
            SHARDING_TEST_MORE_TESTS,
            test(SHARDING_TEST_MORE_TESTS, "testEight")),
        tests);
  }

  private UniqueId test(UniqueId classId, String testName) {
    return classId.append("method", testName + "()");
  }

  private UniqueId testTemplate(UniqueId classId, String testName, String testArgs) {
    return classId.append("test-template", testName + "(" + testArgs + ")");
  }

  private UniqueId testTemplateInvocation(
      UniqueId classId, String testName, String testArgs, int invocation) {
    return testTemplate(classId, testName, testArgs)
        .append("test-template-invocation", "#" + invocation);
  }
}

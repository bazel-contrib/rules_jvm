package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;

import com.github.bazel_contrib.contrib_rules_jvm.junit5.sample.ShardingTestMoreTests;
import com.github.bazel_contrib.contrib_rules_jvm.junit5.sample.ShardingTestTests;
import java.util.Set;
import java.util.stream.Collectors;
import org.junit.Before;
import org.junit.Test;
import org.junit.platform.engine.UniqueId;
import org.junit.platform.engine.discovery.DiscoverySelectors;
import org.junit.platform.launcher.LauncherDiscoveryRequest;
import org.junit.platform.launcher.PostDiscoveryFilter;
import org.junit.platform.launcher.TestIdentifier;
import org.junit.platform.launcher.TestPlan;
import org.junit.platform.launcher.core.LauncherDiscoveryRequestBuilder;
import org.junit.platform.launcher.core.LauncherFactory;

public class ShardingTest {

  private static final UniqueId SHARDING_TEST_TESTS =
      UniqueId.forEngine("junit-jupiter").append("class", ShardingTestTests.class.getName());
  private static final UniqueId SHARDING_TEST_MORE_TESTS =
      UniqueId.forEngine("junit-jupiter").append("class", ShardingTestMoreTests.class.getName());

  private Set<UniqueId> tests;

  @Before
  public void setUp() {
    PostDiscoveryFilter shardFilter = TestSharding.makeShardFilter();
    LauncherDiscoveryRequest request =
        LauncherDiscoveryRequestBuilder.request()
            .selectors(
                DiscoverySelectors.selectPackage(
                    ShardingTest.class.getPackage().getName() + ".sample"))
            .filters(shardFilter)
            .build();
    TestPlan testPlan = LauncherFactory.create().discover(request);
    tests =
        testPlan.getRoots().stream()
            .flatMap(root -> testPlan.getDescendants(root).stream())
            .map(TestIdentifier::getUniqueIdObject)
            .collect(Collectors.toSet());
  }

  @Test
  public void testShard1() {
    assertEquals(
        Set.of(
            SHARDING_TEST_TESTS,
            test(SHARDING_TEST_TESTS, "testSeven"),
            SHARDING_TEST_MORE_TESTS,
            test(SHARDING_TEST_MORE_TESTS, "testTwo"),
            testTemplate(SHARDING_TEST_MORE_TESTS, "testParameterized", "int")),
        tests);
  }

  @Test
  public void testShard2() {
    assertEquals(
        Set.of(
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
            SHARDING_TEST_TESTS,
            test(SHARDING_TEST_TESTS, "testTwo"),
            testTemplate(
                SHARDING_TEST_TESTS, "testRepeated", "org.junit.jupiter.api.RepetitionInfo"),
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
}

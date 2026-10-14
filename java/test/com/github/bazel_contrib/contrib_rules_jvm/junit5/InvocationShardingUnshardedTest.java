package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;

import com.github.bazel_contrib.contrib_rules_jvm.junit5.sample.ShardingAnnotatedTests;
import java.util.HashSet;
import java.util.Set;
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

/**
 * Companion to {@link InvocationShardingTest} which runs without sharding: {@link
 * ShardTemplateInvocations} must have no effect when test sharding is not enabled, so every
 * invocation of every test template runs and nothing is skipped.
 */
public class InvocationShardingUnshardedTest {

  private static final UniqueId ENGINE_ID = UniqueId.forEngine("junit-jupiter");
  private static final UniqueId TEST_CLASS =
      ENGINE_ID.append("class", ShardingAnnotatedTests.class.getName());
  private static final UniqueId SHARDED_PARAMETERIZED =
      TEST_CLASS.append("test-template", "testShardedParameterized(int)");
  private static final UniqueId SHARDED_REPEATED =
      TEST_CLASS.append("test-template", "testShardedRepeated()");
  private static final UniqueId UNSHARDED_PARAMETERIZED =
      TEST_CLASS.append("test-template", "testUnshardedParameterized(int)");

  @Test
  public void allInvocationsRunWhenShardingIsDisabled() {
    assertNull(System.getenv("TEST_TOTAL_SHARDS"));

    LauncherDiscoveryRequest request =
        LauncherDiscoveryRequestBuilder.request()
            .selectors(DiscoverySelectors.selectClass(ShardingAnnotatedTests.class))
            .build();

    Set<UniqueId> started = new HashSet<>();
    Set<UniqueId> skipped = new HashSet<>();
    PostDiscoveryFilter shardFilter = TestSharding.makeShardFilter();
    LauncherConfig config =
        LauncherConfig.builder()
            .addTestExecutionListeners(
                new TestExecutionListener() {
                  @Override
                  public void executionStarted(TestIdentifier testIdentifier) {
                    started.add(testIdentifier.getUniqueIdObject());
                  }

                  @Override
                  public void executionSkipped(TestIdentifier testIdentifier, String reason) {
                    skipped.add(testIdentifier.getUniqueIdObject());
                  }
                })
            .addPostDiscoveryFilters(shardFilter)
            .build();

    LauncherFactory.create(config).execute(request);

    assertEquals(
        Set.of(
            ENGINE_ID,
            TEST_CLASS,
            SHARDED_PARAMETERIZED,
            invocation(SHARDED_PARAMETERIZED, 1),
            invocation(SHARDED_PARAMETERIZED, 2),
            invocation(SHARDED_PARAMETERIZED, 3),
            invocation(SHARDED_PARAMETERIZED, 4),
            invocation(SHARDED_PARAMETERIZED, 5),
            invocation(SHARDED_PARAMETERIZED, 6),
            invocation(SHARDED_PARAMETERIZED, 7),
            SHARDED_REPEATED,
            invocation(SHARDED_REPEATED, 1),
            invocation(SHARDED_REPEATED, 2),
            invocation(SHARDED_REPEATED, 3),
            invocation(SHARDED_REPEATED, 4),
            invocation(SHARDED_REPEATED, 5),
            UNSHARDED_PARAMETERIZED,
            invocation(UNSHARDED_PARAMETERIZED, 1),
            invocation(UNSHARDED_PARAMETERIZED, 2),
            invocation(UNSHARDED_PARAMETERIZED, 3)),
        started);
    assertEquals(Set.of(), skipped);
  }

  private UniqueId invocation(UniqueId templateId, int invocation) {
    return templateId.append("test-template-invocation", "#" + invocation);
  }
}

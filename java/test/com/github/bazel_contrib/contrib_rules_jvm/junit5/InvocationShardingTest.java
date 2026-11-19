package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;

import com.github.bazel_contrib.contrib_rules_jvm.junit5.sample.ShardingAnnotatedTests;
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

/**
 * Tests invocation-level sharding of test templates annotated with {@link
 * ShardTemplateInvocations}.
 *
 * <p>Runs via a Bazel test runner with a known good test sharding implementation and {@code
 * shard_count = 3}, which deterministically runs {@code testShardN} on shard {@code N - 1} (each
 * test method asserts this). Each test method then runs the JUnit Platform launcher in-process with
 * the shard filter installed and asserts exactly which tests were started and which test template
 * invocations were skipped on the current shard:
 *
 * <ul>
 *   <li>Annotated test templates are started on every shard, with started invocations being exactly
 *       the round-robin slice for the shard (invocation N runs on shard (N - 1) % 3) and all other
 *       invocations reported as skipped.
 *   <li>The non-annotated test template keeps the existing behaviour: it is hash-assigned to a
 *       single shard (shard 2 for this test class), where all of its invocations run.
 * </ul>
 */
public class InvocationShardingTest {

  private static final UniqueId ENGINE_ID = UniqueId.forEngine("junit-jupiter");
  private static final UniqueId TEST_CLASS =
      ENGINE_ID.append("class", ShardingAnnotatedTests.class.getName());
  private static final UniqueId SHARDED_PARAMETERIZED =
      TEST_CLASS.append("test-template", "testShardedParameterized(int)");
  private static final UniqueId SHARDED_REPEATED =
      TEST_CLASS.append("test-template", "testShardedRepeated()");
  private static final UniqueId UNSHARDED_PARAMETERIZED =
      TEST_CLASS.append("test-template", "testUnshardedParameterized(int)");

  private Set<UniqueId> started;
  private Set<UniqueId> skipped;

  @Before
  public void setUp() {
    LauncherDiscoveryRequest request =
        LauncherDiscoveryRequestBuilder.request()
            .selectors(DiscoverySelectors.selectClass(ShardingAnnotatedTests.class))
            .build();

    started = new HashSet<>();
    skipped = new HashSet<>();
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
  }

  @Test
  public void testShard1() {
    assertEquals("0", System.getenv("TEST_SHARD_INDEX"));
    assertEquals(
        Set.of(
            ENGINE_ID,
            TEST_CLASS,
            SHARDED_PARAMETERIZED,
            invocation(SHARDED_PARAMETERIZED, 1),
            invocation(SHARDED_PARAMETERIZED, 4),
            invocation(SHARDED_PARAMETERIZED, 7),
            SHARDED_REPEATED,
            invocation(SHARDED_REPEATED, 1),
            invocation(SHARDED_REPEATED, 4)),
        started);
    assertEquals(
        Set.of(
            invocation(SHARDED_PARAMETERIZED, 2),
            invocation(SHARDED_PARAMETERIZED, 3),
            invocation(SHARDED_PARAMETERIZED, 5),
            invocation(SHARDED_PARAMETERIZED, 6),
            invocation(SHARDED_REPEATED, 2),
            invocation(SHARDED_REPEATED, 3),
            invocation(SHARDED_REPEATED, 5)),
        skipped);
  }

  @Test
  public void testShard2() {
    assertEquals("1", System.getenv("TEST_SHARD_INDEX"));
    assertEquals(
        Set.of(
            ENGINE_ID,
            TEST_CLASS,
            SHARDED_PARAMETERIZED,
            invocation(SHARDED_PARAMETERIZED, 2),
            invocation(SHARDED_PARAMETERIZED, 5),
            SHARDED_REPEATED,
            invocation(SHARDED_REPEATED, 2),
            invocation(SHARDED_REPEATED, 5)),
        started);
    assertEquals(
        Set.of(
            invocation(SHARDED_PARAMETERIZED, 1),
            invocation(SHARDED_PARAMETERIZED, 3),
            invocation(SHARDED_PARAMETERIZED, 4),
            invocation(SHARDED_PARAMETERIZED, 6),
            invocation(SHARDED_PARAMETERIZED, 7),
            invocation(SHARDED_REPEATED, 1),
            invocation(SHARDED_REPEATED, 3),
            invocation(SHARDED_REPEATED, 4)),
        skipped);
  }

  @Test
  public void testShard3() {
    assertEquals("2", System.getenv("TEST_SHARD_INDEX"));
    assertEquals(
        Set.of(
            ENGINE_ID,
            TEST_CLASS,
            SHARDED_PARAMETERIZED,
            invocation(SHARDED_PARAMETERIZED, 3),
            invocation(SHARDED_PARAMETERIZED, 6),
            SHARDED_REPEATED,
            invocation(SHARDED_REPEATED, 3),
            UNSHARDED_PARAMETERIZED,
            invocation(UNSHARDED_PARAMETERIZED, 1),
            invocation(UNSHARDED_PARAMETERIZED, 2),
            invocation(UNSHARDED_PARAMETERIZED, 3)),
        started);
    assertEquals(
        Set.of(
            invocation(SHARDED_PARAMETERIZED, 1),
            invocation(SHARDED_PARAMETERIZED, 2),
            invocation(SHARDED_PARAMETERIZED, 4),
            invocation(SHARDED_PARAMETERIZED, 5),
            invocation(SHARDED_PARAMETERIZED, 7),
            invocation(SHARDED_REPEATED, 1),
            invocation(SHARDED_REPEATED, 2),
            invocation(SHARDED_REPEATED, 4),
            invocation(SHARDED_REPEATED, 5)),
        skipped);
  }

  private UniqueId invocation(UniqueId templateId, int invocation) {
    return templateId.append("test-template-invocation", "#" + invocation);
  }
}

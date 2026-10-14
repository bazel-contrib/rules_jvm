package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.extension.ConditionEvaluationResult;
import org.junit.jupiter.api.extension.ExecutionCondition;
import org.junit.jupiter.api.extension.ExtensionContext;
import org.junit.platform.engine.UniqueId;

/**
 * An {@link ExecutionCondition} which enables each invocation of a test template (such as a
 * {@code @ParameterizedTest} or {@code @RepeatedTest}) on exactly one Bazel test shard.
 *
 * <p>This condition is registered by annotating a test template method (or its class) with {@link
 * ShardTemplateInvocations}. It cooperates with the {@code PostDiscoveryFilter} created by {@link
 * TestSharding#makeShardFilter()}: the filter includes annotated test templates on every shard
 * (rather than hashing them onto a single shard), and this condition then enables each invocation
 * on exactly one shard, assigning invocation {@code N} (1-based) round-robin to shard {@code (N -
 * 1) % TEST_TOTAL_SHARDS}.
 */
public class TemplateInvocationShardingCondition implements ExecutionCondition {

  private static final String TEST_TEMPLATE_INVOCATION_SEGMENT_TYPE = "test-template-invocation";

  @Override
  public ConditionEvaluationResult evaluateExecutionCondition(ExtensionContext context) {
    if (!TestSharding.isShardingEnabled()) {
      return ConditionEvaluationResult.enabled("test sharding is disabled");
    }

    UniqueId.Segment lastSegment = UniqueId.parse(context.getUniqueId()).getLastSegment();
    if (!TEST_TEMPLATE_INVOCATION_SEGMENT_TYPE.equals(lastSegment.getType())) {
      // Containers (the test class, the test template itself, ...) must always be enabled: the
      // per-invocation decision is made when each invocation's own context is evaluated.
      return ConditionEvaluationResult.enabled("not a test template invocation");
    }

    int invocation;
    try {
      // The segment value has the form "#N" where N is the 1-based invocation index.
      invocation = Integer.parseInt(lastSegment.getValue().substring(1));
    } catch (RuntimeException e) {
      // If the invocation index cannot be determined, prefer running the invocation on every
      // shard (duplicated work) over silently dropping it from all shards.
      return ConditionEvaluationResult.enabled(
          "unable to parse test template invocation index from segment value '"
              + lastSegment.getValue()
              + "'");
    }

    if (Math.floorMod(invocation - 1, TestSharding.getTotalShards())
        == TestSharding.getShardIndex()) {
      return ConditionEvaluationResult.enabled("test template invocation is in current shard");
    }
    return ConditionEvaluationResult.disabled("test template invocation is in different shard");
  }
}

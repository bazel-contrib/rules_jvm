package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.lang.annotation.Documented;
import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;
import org.junit.jupiter.api.extension.ExtendWith;

/**
 * Opt-in annotation that distributes the individual invocations of a JUnit Jupiter test template
 * (for example {@code @ParameterizedTest} or {@code @RepeatedTest}) across Bazel test shards.
 *
 * <p>By default the shard filter treats a test template as a single unit, because the invocations
 * of a template do not exist at discovery time, which is when the filter runs. As a result, all
 * invocations of a parameterized or repeated test run on the same shard. Annotating the template
 * method (or its class, to opt in every template in the class) instead includes the template on
 * every shard and enables each invocation on exactly one shard: invocation {@code N} (1-based) is
 * assigned round-robin to shard {@code (N - 1) % TEST_TOTAL_SHARDS}, which is deterministic and
 * keeps per-shard invocation counts balanced. Invocations assigned to other shards are reported as
 * skipped on the current shard.
 *
 * <p>This requires the set and order of invocations to be stable across shards: every shard must
 * generate the same invocations in the same order, so argument providers must be deterministic. If
 * they are not, some invocations may run on multiple shards or not at all.
 *
 * <p>This annotation is not supported on {@code @TestFactory} methods: dynamic tests cannot be
 * conditionally disabled, so test factories continue to be sharded as a single unit.
 *
 * <p>Has no effect when test sharding is not enabled.
 */
@Target({ElementType.METHOD, ElementType.TYPE})
@Retention(RetentionPolicy.RUNTIME)
@Documented
@ExtendWith(TemplateInvocationShardingCondition.class)
public @interface ShardTemplateInvocations {}

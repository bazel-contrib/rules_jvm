package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.file.FileAlreadyExistsException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.attribute.FileTime;
import org.junit.platform.engine.FilterResult;
import org.junit.platform.launcher.PostDiscoveryFilter;

final class TestSharding {

  public static PostDiscoveryFilter makeShardFilter() {
    if (!isShardingEnabled()) {
      return testDescriptor -> FilterResult.included("test sharding disabled");
    }

    // Let Bazel know that this runner supports sharding.
    // https://bazel.build/reference/test-encyclopedia#test-sharding
    touchShardFile();

    final long totalShards = getTotalShards();

    return testDescriptor -> {
      // We want to filter on the level of actual tests, not contained. Since PostDiscoveryFilters
      // do not see dynamic tests and invocations obtained from test templates, we make an exception
      // and filter their container representations instead.
      if (!testDescriptor.isTest() && !testDescriptor.mayRegisterTests()) {
        return FilterResult.included("non-test nodes in the test plan are always included");
      }

      // A JUnit test plan has a tree structure with potentially multiple roots that can change at
      // execution time due to dynamic tests. Rather than enumerating all tests and assigning them
      // to shards, use a hash of the test ID.
      //
      // Use Math.floorMod instead of % to ensure a positive result even if the hash is negative.
      long shard = Math.floorMod(testDescriptor.getUniqueId().hashCode(), totalShards);

      if (shard == getShardIndex()) {
        return FilterResult.included("test is in current shard");
      } else {
        return FilterResult.excluded("test is in different shard");
      }
    };
  }

  private static boolean isShardingEnabled() {
    return System.getenv("TEST_TOTAL_SHARDS") != null;
  }

  private static long getShardIndex() {
    return Integer.parseUnsignedInt(System.getenv().getOrDefault("TEST_SHARD_INDEX", "0"));
  }

  private static long getTotalShards() {
    return Integer.parseUnsignedInt(System.getenv().getOrDefault("TEST_TOTAL_SHARDS", "1"));
  }

  private static void touchShardFile() {
    String shardStatusPath = System.getenv("TEST_SHARD_STATUS_FILE");
    if (shardStatusPath == null) {
      return;
    }
    Path shardFile = Paths.get(shardStatusPath);
    try {
      touch(shardFile);
    } catch (IOException e) {
      throw new UncheckedIOException("Failed to touch shard status file " + shardFile, e);
    }
  }

  private static void touch(Path file) throws IOException {
    try {
      Files.createFile(file);
    } catch (FileAlreadyExistsException e) {
      Files.setLastModifiedTime(file, FileTime.fromMillis(System.currentTimeMillis()));
    }
  }

  private TestSharding() {}
}

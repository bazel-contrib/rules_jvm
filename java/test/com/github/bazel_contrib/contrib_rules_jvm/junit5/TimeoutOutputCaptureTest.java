package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.Test;

/**
 * Integration test that intentionally times out to verify stdout capture.
 *
 * <p>Run with: bazel test //java/test/.../junit5:timeout-output-capture-test --test_timeout=3
 *
 * <p>Then check test.xml for system-out containing the log lines.
 *
 * <p>This test is marked manual so it doesn't run in normal test suites.
 */
public class TimeoutOutputCaptureTest {

  @Test
  void testOutputCapturedOnTimeout() throws InterruptedException {
    System.out.println("[TIMEOUT-TEST] Starting test");
    System.out.println("[TIMEOUT-TEST] Line 1");
    System.out.println("[TIMEOUT-TEST] Line 2");
    System.out.flush();

    // Sleep longer than the test timeout
    Thread.sleep(60_000);

    // This should never be reached
    System.out.println("[TIMEOUT-TEST] ERROR: Test completed without timeout!");
  }
}

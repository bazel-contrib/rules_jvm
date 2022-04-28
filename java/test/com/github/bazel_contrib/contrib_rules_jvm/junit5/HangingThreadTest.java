package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.util.concurrent.atomic.AtomicBoolean;
import org.junit.jupiter.api.Test;

public class HangingThreadTest {

  // This thread isn't a daemon thread, so after the test is complete
  // the JVM will wait for it to exit. However, all the tests are now
  // done, so the junit runner should exit cleanly.
  private static final AtomicBoolean STARTED = new AtomicBoolean(false);
  private static final Thread SNAGGING_THREAD =
      new Thread(
          () -> {
            STARTED.set(true);
            try {
              Thread.sleep(Long.MAX_VALUE);
            } catch (InterruptedException e) {
              // Swallow
            }
          });

  static {
    SNAGGING_THREAD.start();
  }

  // We need this to get the thread to start running.
  @Test
  public void hangForever() throws InterruptedException {
    while (!STARTED.get()) {
      Thread.sleep(100);
    }
  }
}

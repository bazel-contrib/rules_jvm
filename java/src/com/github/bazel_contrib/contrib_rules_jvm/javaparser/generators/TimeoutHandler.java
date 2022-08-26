package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import com.google.errorprone.annotations.concurrent.GuardedBy;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class TimeoutHandler {
  private static final Logger logger = LoggerFactory.getLogger(TimeoutHandler.class);

  private final ScheduledExecutorService executor;
  private final int timeoutSeconds;
  private final AtomicInteger inFlightRequests = new AtomicInteger(0);

  private final Object lastFutureLock = new Object();

  @GuardedBy("lastFutureLock")
  private ScheduledFuture<?> lastFuture = null;

  public TimeoutHandler(ScheduledExecutorService executor, int timeoutSeconds) {
    this.executor = executor;
    this.timeoutSeconds = timeoutSeconds;
    schedule();
  }

  public void startedRequest() {
    this.inFlightRequests.getAndIncrement();
  }

  public void finishedRequest() {
    this.inFlightRequests.getAndDecrement();
    schedule();
  }

  void schedule() {
    if (this.timeoutSeconds <= 0) {
      return;
    }

    synchronized (lastFutureLock) {
      ScheduledFuture<?> last = this.lastFuture;
      if (last != null) {
        last.cancel(true);
      }
      this.lastFuture =
          this.executor.schedule(
              () -> {
                if (inFlightRequests.get() == 0) {
                  logger.debug(
                      "Saw no requests in flight after {} seconds, terminating.", timeoutSeconds);
                  System.exit(0);
                }
              },
              timeoutSeconds,
              TimeUnit.SECONDS);
    }
  }

  public void cancelOutstanding() {
    synchronized (lastFutureLock) {
      if (lastFuture != null) {
        lastFuture.cancel(true);
        lastFuture = null;
      }
    }
    this.executor.shutdownNow();
  }
}

package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import com.google.errorprone.annotations.concurrent.GuardedBy;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.ScheduledThreadPoolExecutor;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class TimeoutHandler {
  private static final Logger logger = LoggerFactory.getLogger(TimeoutHandler.class);

  // Anything which shuts down the executor, or relies on it not being shutdown, must do so under
  // this lock.
  @GuardedBy("lastFutureLock")
  private final ScheduledExecutorService executor;

  private final int timeoutSeconds;
  private final AtomicInteger inFlightRequests = new AtomicInteger(0);

  private final Object lastFutureLock = new Object();

  @GuardedBy("lastFutureLock")
  private ScheduledFuture<?> lastFuture = null;

  public TimeoutHandler(int timeoutSeconds) {
    this.executor = new ScheduledThreadPoolExecutor(1);
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
      if (this.executor.isShutdown()) {
        // If the executor is already shutdown, the process is already terminating - we neither can
        // (because it's shutdown) nor need to (because we're about to terminate) schedule a new
        // task.
        return;
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

  /**
   * Cancel any outstanding scheduled tasks, and shut down any background threads.
   *
   * <p>This object becomes useless after this method is called - it will not perform its timeout
   * functionality any more.
   */
  public void cancelOutstandingAndStopScheduling() {
    synchronized (lastFutureLock) {
      if (lastFuture != null) {
        lastFuture.cancel(true);
        lastFuture = null;
      }
      this.executor.shutdownNow();
    }
  }
}

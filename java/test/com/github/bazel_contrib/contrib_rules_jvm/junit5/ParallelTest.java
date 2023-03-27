package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.util.Collections;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.logging.Handler;
import java.util.logging.Level;
import java.util.logging.LogManager;
import java.util.logging.LogRecord;
import java.util.stream.Collectors;
import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.RepeatedTest;

/**
 * Verifies that {@link BazelJUnitOutputListener} does not throw any exceptions during parallel test
 * execution (in particular, no {@link java.util.ConcurrentModificationException} by asserting that
 * JUnit logs no warnings - every exception thrown by a {@link
 * org.junit.platform.launcher.TestExecutionListener} results in one.
 */
public class ParallelTest {

  private static LogHandler logHandler;

  public static class LogHandler extends Handler {

    Set<LogRecord> warnings = Collections.newSetFromMap(new ConcurrentHashMap<>());

    @Override
    public void publish(LogRecord logRecord) {
      if (logRecord.getLevel().intValue() >= Level.WARNING.intValue()) {
        warnings.add(logRecord);
      }
    }

    public void assertNoWarnings() {
      assertEquals(
          Collections.emptyList(),
          warnings.stream().map(LogRecord::getMessage).collect(Collectors.toList()));
    }

    @Override
    public void flush() {}

    @Override
    public void close() throws SecurityException {}
  }

  @BeforeAll
  static void registerLogHandler() {
    logHandler = new LogHandler();
    LogManager.getLogManager().getLogger("").addHandler(logHandler);
  }

  @RepeatedTest(1000)
  void testInParallel() {}

  @AfterAll
  static void assertNoWarnings() {
    logHandler.assertNoWarnings();
  }
}

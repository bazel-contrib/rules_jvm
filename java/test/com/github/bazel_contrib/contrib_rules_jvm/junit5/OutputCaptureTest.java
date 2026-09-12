package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static java.nio.charset.StandardCharsets.UTF_8;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertSame;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.ByteArrayOutputStream;
import java.io.PrintStream;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class OutputCaptureTest {

  private PrintStream originalOut;
  private PrintStream originalErr;

  @BeforeEach
  void setUp() {
    originalOut = System.out;
    originalErr = System.err;
  }

  @AfterEach
  void tearDown() {
    System.setOut(originalOut);
    System.setErr(originalErr);
  }

  @Test
  void testCapturesStdout() {
    OutputCapture capture = new OutputCapture();
    capture.start();

    System.out.print("hello");

    capture.stop();

    assertTrue(capture.hasStdout());
    assertEquals("hello", capture.getStdout());
  }

  @Test
  void testCapturesStderr() {
    OutputCapture capture = new OutputCapture();
    capture.start();

    System.err.print("error message");

    capture.stop();

    assertTrue(capture.hasStderr());
    assertEquals("error message", capture.getStderr());
  }

  @Test
  void testRestoresOriginalStreams() {
    // Use assertSame with inline references to avoid PMD CloseResource false positive
    OutputCapture capture = new OutputCapture();
    assertSame(originalOut, System.out, "stdout should match before start");
    assertSame(originalErr, System.err, "stderr should match before start");

    capture.start();
    capture.stop();

    assertSame(originalOut, System.out, "stdout should be restored after stop");
    assertSame(originalErr, System.err, "stderr should be restored after stop");
  }

  @Test
  void testIdempotentStartStop() {
    OutputCapture capture = new OutputCapture();

    // Multiple starts should be safe
    capture.start();
    capture.start();

    System.out.print("test");

    // Multiple stops should be safe
    capture.stop();
    capture.stop();

    assertEquals("test", capture.getStdout());
  }

  @Test
  void testEmptyCapture() {
    OutputCapture capture = new OutputCapture();
    capture.start();
    capture.stop();

    assertFalse(capture.hasStdout());
    assertFalse(capture.hasStderr());
    assertEquals("", capture.getStdout());
    assertEquals("", capture.getStderr());
  }

  @Test
  void testOutputStillGoesToOriginalStream() {
    // Capture the original stdout to a buffer
    ByteArrayOutputStream originalBuffer = new ByteArrayOutputStream();
    PrintStream testOut = new PrintStream(originalBuffer, true, UTF_8);
    System.setOut(testOut);

    OutputCapture capture = new OutputCapture();
    capture.start();

    System.out.print("dual output");

    capture.stop();

    // Both our capture and the "original" (our test buffer) should have the output
    assertEquals("dual output", capture.getStdout());
    assertEquals("dual output", originalBuffer.toString(UTF_8));
  }
}

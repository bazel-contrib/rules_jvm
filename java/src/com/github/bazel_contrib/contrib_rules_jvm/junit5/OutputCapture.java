package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.OutputStream;
import java.io.PrintStream;
import java.nio.charset.StandardCharsets;

/**
 * Captures stdout and stderr independently of JUnit Platform's capture mechanism.
 *
 * <p>This class wraps System.out and System.err with tee streams that write to both the original
 * stream and an internal buffer. This ensures output is captured even when tests are interrupted
 * (e.g., by SIGTERM due to timeout) before JUnit Platform can publish the captured output.
 */
public class OutputCapture {

  private static final String MAX_BYTES_PROPERTY = "junit5.outputCapture.maxBytes";
  private static final int DEFAULT_MAX_BYTES = 10 * 1024 * 1024; // 10MB

  private final PrintStream originalOut;
  private final PrintStream originalErr;
  private final BoundedByteArrayOutputStream stdoutBuffer;
  private final BoundedByteArrayOutputStream stderrBuffer;
  private volatile boolean started = false;

  public OutputCapture() {
    this.originalOut = System.out;
    this.originalErr = System.err;

    int maxBytes = getMaxCaptureSize();
    this.stdoutBuffer = new BoundedByteArrayOutputStream(maxBytes);
    this.stderrBuffer = new BoundedByteArrayOutputStream(maxBytes);
  }

  private static int getMaxCaptureSize() {
    return Integer.getInteger(MAX_BYTES_PROPERTY, DEFAULT_MAX_BYTES);
  }

  /** Start capturing stdout and stderr. */
  public synchronized void start() {
    if (started) {
      return;
    }

    System.setOut(new PrintStream(new TeeOutputStream(originalOut, stdoutBuffer), true));
    System.setErr(new PrintStream(new TeeOutputStream(originalErr, stderrBuffer), true));
    started = true;
  }

  /** Stop capturing and restore original streams. */
  public synchronized void stop() {
    if (!started) {
      return;
    }

    System.setOut(originalOut);
    System.setErr(originalErr);
    started = false;
  }

  /** Get captured stdout content. */
  public String getStdout() {
    return stdoutBuffer.toString(StandardCharsets.UTF_8);
  }

  /** Get captured stderr content. */
  public String getStderr() {
    return stderrBuffer.toString(StandardCharsets.UTF_8);
  }

  /** Check if any stdout was captured. */
  public boolean hasStdout() {
    return stdoutBuffer.size() > 0;
  }

  /** Check if any stderr was captured. */
  public boolean hasStderr() {
    return stderrBuffer.size() > 0;
  }

  /** OutputStream that writes to two underlying streams. */
  private static class TeeOutputStream extends OutputStream {
    private final OutputStream primary;
    private final OutputStream secondary;

    TeeOutputStream(OutputStream primary, OutputStream secondary) {
      this.primary = primary;
      this.secondary = secondary;
    }

    @Override
    public void write(int b) throws IOException {
      primary.write(b);
      secondary.write(b);
    }

    @Override
    public void write(byte[] b) throws IOException {
      primary.write(b);
      secondary.write(b);
    }

    @Override
    public void write(byte[] b, int off, int len) throws IOException {
      primary.write(b, off, len);
      secondary.write(b, off, len);
    }

    @Override
    public void flush() throws IOException {
      primary.flush();
      secondary.flush();
    }

    @Override
    public void close() throws IOException {
      primary.flush();
      secondary.flush();
    }
  }

  /** ByteArrayOutputStream with a maximum size limit to prevent OOM. */
  private static class BoundedByteArrayOutputStream extends ByteArrayOutputStream {
    private final int maxSize;
    private boolean overflow = false;

    BoundedByteArrayOutputStream(int maxSize) {
      this.maxSize = maxSize;
    }

    @Override
    public synchronized void write(int b) {
      if (size() >= maxSize) {
        overflow = true;
        return;
      }
      super.write(b);
    }

    @Override
    public synchronized void write(byte[] b, int off, int len) {
      if (size() >= maxSize) {
        overflow = true;
        return;
      }

      int available = maxSize - size();
      int toWrite = Math.min(len, available);
      super.write(b, off, toWrite);

      if (toWrite < len) {
        overflow = true;
      }
    }

    @Override
    public synchronized String toString(java.nio.charset.Charset charset) {
      String result = new String(toByteArray(), charset);
      if (overflow) {
        result += "\n[OUTPUT TRUNCATED - exceeded " + (maxSize / 1024 / 1024) + "MB limit]";
      }
      return result;
    }
  }
}

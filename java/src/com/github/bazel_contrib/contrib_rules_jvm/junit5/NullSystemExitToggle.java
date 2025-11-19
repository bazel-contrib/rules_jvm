package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.util.Optional;

public class NullSystemExitToggle implements SystemExitToggle {

  private Optional<Thread> shutdownHook;

  public NullSystemExitToggle() {
    this.shutdownHook = Optional.empty();
  }

  @Override
  public void prevent() {
    if (!shutdownHook.isEmpty()) {
      throw new RuntimeException("SystemExitDetectingShutdownHook already added");
    }
    // Install a shutdown hook so people can track down what went wrong
    // if a test calls `System.exit`
    Thread shutdownHook = SystemExitDetectingShutdownHook.newShutdownHook(System.err);
    this.shutdownHook = Optional.of(shutdownHook);
    Runtime.getRuntime().addShutdownHook(shutdownHook);
  }

  @Override
  public void allow() {
    if (shutdownHook.isEmpty()) {
      throw new RuntimeException("SystemExitDetectingShutdownHook never added");
    }
    Runtime.getRuntime().removeShutdownHook(shutdownHook.get());
  }
}

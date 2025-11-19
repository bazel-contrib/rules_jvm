package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertThrows;

import org.junit.jupiter.api.Test;

public class NullSystemExitToggleTest {

  @Test
  public void prevent_alreadyPrevented_throwsRuntimeException() {
    // Arrange.
    NullSystemExitToggle toggle = new NullSystemExitToggle();

    // Act and Assert.
    toggle.prevent();
    assertThrows(RuntimeException.class, () -> toggle.prevent());

    // Cleanup.
    toggle.allow();
  }

  @Test
  public void allow_alreadyAllowed_throwsRuntimeException() {
    // Arrange.
    NullSystemExitToggle toggle = new NullSystemExitToggle();

    // Act and Assert.
    assertThrows(RuntimeException.class, () -> toggle.allow());
  }

  @Test
  public void normalFlow_success() {
    // Arrange.
    NullSystemExitToggle toggle = new NullSystemExitToggle();

    // Act and (implicit) Assert.
    try {
      toggle.prevent();
    } finally {
      toggle.allow();
    }
  }
}

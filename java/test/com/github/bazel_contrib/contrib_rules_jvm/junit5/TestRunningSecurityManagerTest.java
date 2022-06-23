package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertThrows;

import java.security.Permission;

import edu.umd.cs.findbugs.annotations.SuppressFBWarnings;
import org.junit.jupiter.api.Test;

@SuppressFBWarnings("THROWS_METHOD_THROWS_CLAUSE_THROWABLE")
public class TestRunningSecurityManagerTest {

  @Test
  void shouldStifleSystemExitCalls() {
    SecurityManager sm = new TestRunningSecurityManager(null);
    assertThrows(SecurityException.class, () -> sm.checkExit(2));
  }

  @Test
  void shouldDelegateToExistingSecurityManagerIfPresent() {
    SecurityManager permissive = new TestRunningSecurityManager(null);
    Permission permission = new RuntimePermission("example.permission");
    SecurityManager restrictive =
        new TestRunningSecurityManager(
            new SecurityManager() {
              @Override
              public void checkPermission(Permission perm) {
                if (permission == perm) {
                  throw new SecurityException("Oh noes!");
                }
              }
            });

    // This should do nothing, but if an exception is thrown, our test fails.
    permissive.checkPermission(permission);

    // Whereas this delegates down to the custom security manager
    assertThrows(SecurityException.class, () -> restrictive.checkPermission(permission));
  }
}

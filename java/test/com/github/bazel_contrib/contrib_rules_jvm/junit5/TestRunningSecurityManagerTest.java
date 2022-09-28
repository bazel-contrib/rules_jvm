package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertThrows;

import edu.umd.cs.findbugs.annotations.SuppressFBWarnings;
import java.security.Permission;
import org.junit.jupiter.api.Test;

@SuppressFBWarnings("THROWS_METHOD_THROWS_CLAUSE_THROWABLE")
public class TestRunningSecurityManagerTest {

  @Test
  void shouldStifleSystemExitCalls() {
    var sm = new TestRunningSecurityManager();
    sm.setDelegateSecurityManager(null);
    assertThrows(SecurityException.class, () -> sm.checkExit(2));
  }

  @Test
  void shouldDelegateToExistingSecurityManagerIfPresent() {
    SecurityManager permissive = new TestRunningSecurityManager();
    Permission permission = new RuntimePermission("example.permission");

    var restrictive = new TestRunningSecurityManager();
    restrictive.setDelegateSecurityManager(
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

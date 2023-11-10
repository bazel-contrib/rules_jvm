package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.security.Permission;

public class TestRunningSecurityManager extends SecurityManager {
  private static final RuntimePermission SET_SECURITY_MANAGER_PERMISSION =
      new RuntimePermission("setSecurityManager");
  private boolean allowExitCall = false;
  private SecurityManager delegateSecurityManager;

  public void setDelegateSecurityManager(SecurityManager securityManager) {
    this.delegateSecurityManager = securityManager;
  }

  void allowExitCall() {
    allowExitCall = true;
  }

  @Override
  public void checkExit(int status) {
    if (!allowExitCall) {
      throw new SecurityException("Attempt to call System.exit");
    }
  }

  @Override
  public void checkPermission(Permission perm) {
    if (SET_SECURITY_MANAGER_PERMISSION.equals(perm)) {
      if (allowExitCall) {
        return;
      }
      if (System.getProperty("bazel.junit5runner.allowSettingSecurityManager") != null) {
        System.err.println(
            "Warning: junit runner security manager replaced, calls to System.exit will not be"
                + " blocked");
      } else {
        throw new SecurityException("Replacing the security manager is not allowed");
      }
    }

    if (delegateSecurityManager != null) {
      delegateSecurityManager.checkPermission(perm);
    }
  }

  @Override
  public void checkPermission(Permission perm, Object context) {
    // The default implementation of the SecurityManager checks to see
    // if the `context` is an `AccessControlContext`, and if it is calls
    // `checkPermission` on that. However, when there's no security
    // manager installed, there's never a problem. We're going to pretend
    // that we are "no security manager" installed and just allow things
    // to happen because that's how most people are running their tests.

    if (delegateSecurityManager != null) {
      delegateSecurityManager.checkPermission(perm, context);
    }
  }
}

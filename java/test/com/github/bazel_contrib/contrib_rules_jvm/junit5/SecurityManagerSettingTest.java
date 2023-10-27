package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.security.Permission;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertThrows;

import edu.umd.cs.findbugs.annotations.SuppressFBWarnings;

@SuppressFBWarnings("THROWS_METHOD_THROWS_CLAUSE_THROWABLE")
public class SecurityManagerSettingTest {

  private static final String ALLOW_SETTING_SECURITY_MANAGER_PROPERTY = "bazel.junit5runner.allowSettingSecurityManager";

  @Test
  void testCanSetSecurityManagerWhenPropertyIsTrue() {
    System.setProperty(ALLOW_SETTING_SECURITY_MANAGER_PROPERTY, "true");
    SecurityManager originalSecurityManager = System.getSecurityManager();
    SecurityManager testSecurityManager = new SecurityManager() {
      @Override
       public void checkPermission(Permission perm) {}
    };

    try {
      System.setSecurityManager(testSecurityManager);
      assertEquals(testSecurityManager, System.getSecurityManager());
    } finally {
      System.setSecurityManager(originalSecurityManager);
      System.clearProperty(ALLOW_SETTING_SECURITY_MANAGER_PROPERTY);
    }
  }

  @Test
  void testCannotSetSecurityManagerWhenPropertyIsNotSet() {
    assertNull(System.getProperty(ALLOW_SETTING_SECURITY_MANAGER_PROPERTY));
    assertThrows(SecurityException.class, () -> System.setSecurityManager(new SecurityManager()));
  }
}
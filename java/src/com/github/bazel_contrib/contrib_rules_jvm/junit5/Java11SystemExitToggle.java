package com.github.bazel_contrib.contrib_rules_jvm.junit5;

public class Java11SystemExitToggle implements SystemExitToggle {

  private SecurityManager previousSecurityManager;
  private TestRunningSecurityManager testingSecurityManager;

  @Override
  public void prevent() {
    previousSecurityManager = System.getSecurityManager();
    testingSecurityManager = new TestRunningSecurityManager();
    testingSecurityManager.setDelegateSecurityManager(previousSecurityManager);

    System.setSecurityManager(testingSecurityManager);
  }

  @Override
  public void allow() {
    testingSecurityManager.allowExitCall();
    System.setSecurityManager(testingSecurityManager);
    testingSecurityManager = null;
    previousSecurityManager = null;
  }
}

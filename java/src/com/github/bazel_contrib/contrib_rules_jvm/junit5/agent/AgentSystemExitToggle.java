package com.github.bazel_contrib.contrib_rules_jvm.junit5.agent;

import com.github.bazel_contrib.contrib_rules_jvm.junit5.SystemExitToggle;

public class AgentSystemExitToggle implements SystemExitToggle {

  @Override
  public void prevent() {
    SystemExitAgent.preventSystemExit();
  }

  @Override
  public void allow() {
    SystemExitAgent.allowSystemExit();
  }
}

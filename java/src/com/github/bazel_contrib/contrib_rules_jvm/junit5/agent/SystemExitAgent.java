package com.github.bazel_contrib.contrib_rules_jvm.junit5.agent;

import java.lang.instrument.ClassFileTransformer;
import java.lang.instrument.Instrumentation;
import java.net.URL;
import java.net.URLClassLoader;

public class SystemExitAgent {

  static final String PREVENT_EXIT_PROPERTY = "bazel.junit5runner.preventSystemExit";

  private static final String TRANSFORMER_CLASS =
      "com.github.bazel_contrib.contrib_rules_jvm.junit5.agent.RuntimeExitTransformer";

  public static void premain(String args, Instrumentation inst) {
    try {
      if (args == null || args.isEmpty()) {
        System.err.println("system-exit-agent requires the agent jar path as an argument");
        return;
      }

      // The agent jar path is passed as an argument via -javaagent:path=path
      URL agentJarUrl = new java.io.File(args).toURI().toURL();

      // Create an isolated classloader with null parent (only bootstrap classloader).
      // This ensures ASM classes are not visible to the application classpath.
      URLClassLoader isolatedLoader = new URLClassLoader(new URL[] {agentJarUrl}, null);

      Class<?> transformerClass = isolatedLoader.loadClass(TRANSFORMER_CLASS);
      ClassFileTransformer transformer =
          (ClassFileTransformer) transformerClass.getConstructor().newInstance();

      inst.addTransformer(transformer, true);
      inst.retransformClasses(Runtime.class);
    } catch (Exception e) {
      System.err.println("Failed to initialize system exit agent: " + e.getMessage());
      e.printStackTrace(System.err);
    }
  }

  public static void preventSystemExit() {
    System.setProperty(PREVENT_EXIT_PROPERTY, "true");
  }

  public static void allowSystemExit() {
    System.clearProperty(PREVENT_EXIT_PROPERTY);
  }
}

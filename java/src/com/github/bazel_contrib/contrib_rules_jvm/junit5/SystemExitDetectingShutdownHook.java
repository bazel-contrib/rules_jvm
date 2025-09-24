// Copyright 2024 The Bazel Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package com.google.testing.junit.runner.internal;

// This code has been lifted from:
// https://github.com/bazelbuild/bazel/blob/a74b12a652ba9b28a078e4270bc144973e26b2a1/src/java_tools/junitrunner/java/com/google/testing/junit/runner/internal/SystemExitDetectingShutdownHook.java#L27

package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.io.PrintStream;
import java.util.ArrayList;
import java.util.List;

public class SystemExitDetectingShutdownHook {

  private SystemExitDetectingShutdownHook() {}

  public static Thread newShutdownHook(PrintStream testRunnerOut) {
    Runnable hook =
        () -> {
          boolean foundRuntimeExit = false;
          for (StackTraceElement[] stack : Thread.getAllStackTraces().values()) {
            List<String> framesStartingWithRuntimeExit = new ArrayList<>();
            boolean foundRuntimeExitInThisThread = false;
            for (StackTraceElement frame : stack) {
              if (!foundRuntimeExitInThisThread
                  && frame.getClassName().equals("java.lang.Runtime")
                  && frame.getMethodName().equals("exit")) {
                foundRuntimeExitInThisThread = true;
              }
              if (foundRuntimeExitInThisThread) {
                framesStartingWithRuntimeExit.add(frameString(frame));
              }
            }
            if (foundRuntimeExitInThisThread) {
              foundRuntimeExit = true;
              testRunnerOut.println("\nSystem.exit or Runtime.exit was called!");
              testRunnerOut.println(String.join("\n", framesStartingWithRuntimeExit));
            }
          }
          if (foundRuntimeExit) {
            // We must call halt rather than exit, because exit would lead to a deadlock. We use a
            // hopefully unique exit code to make it easier to identify this case.
            Runtime.getRuntime().halt(121);
          }
        };
    return new Thread(hook, "SystemExitDetectingShutdownHook");
  }

  private static String frameString(StackTraceElement frame) {
    return String.format(
        "        at %s.%s(%s:%d)",
        frame.getClassName(), frame.getMethodName(), frame.getFileName(), frame.getLineNumber());
  }
}

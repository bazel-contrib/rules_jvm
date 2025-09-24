package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import edu.umd.cs.findbugs.annotations.SuppressFBWarnings;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

public class NestedClassesTest {
  static class First {
    @Test
    public void shouldBeExecuted() {
      System.out.println(">>>> Executed test in NestedClassesTest$First");
    }
  }

  @Nested
  @SuppressFBWarnings("SIC_INNER_SHOULD_BE_STATIC")
  class Second {
    @Test
    public void shouldBeExecuted() {
      System.out.println(">>>> Executed test in NestedClassesTest$Second");
    }

    @Nested
    @SuppressFBWarnings("SIC_INNER_SHOULD_BE_STATIC")
    class Third {
      @Test
      public void shouldBeExecuted() {
        System.out.println(">>>> Executed test in NestedClassesTest$Second$Third");
      }

      @Nested
      @SuppressFBWarnings("SIC_INNER_SHOULD_BE_STATIC")
      class Fourth {
        @Test
        public void shouldBeExecuted() {
          System.out.println(">>>> Executed test in NestedClassesTest$Second$Third$Fourth");
        }

        @Nested
        @SuppressFBWarnings("SIC_INNER_SHOULD_BE_STATIC")
        class Fifth {
          @Test
          public void shouldBeExecuted() {
            System.out.println(">>>> Executed test in NestedClassesTest$Second$Third$Fourth$Fifth");
          }
        }
      }
    }
  }

  @Test
  public void shouldBeExecuted() {
    System.out.println(">>>> Executed test in NestedClassesTest");
  }
}

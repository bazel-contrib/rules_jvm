package com.github.bazel_contrib.contrib_rules_jvm.comparative;

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
  class Second {
    @Test
    public void shouldBeExecuted() {
      System.out.println(">>>> Executed test in NestedClassesTest$Second");
    }

    @Nested
    class Third {
      @Test
      public void shouldBeExecuted() {
        System.out.println(">>>> Executed test in NestedClassesTest$Second$Third");
      }

      @Nested
      class Fourth {
        @Test
        public void shouldBeExecuted() {
          System.out.println(">>>> Executed test in NestedClassesTest$Second$Third$Fourth");
        }

        @Nested
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

package com.github.bazel_contrib.contrib_rules_jvm.junit5.select_deps;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertTrue;

/** Test that verifies java_test_suite works with select() in deps. */
class SelectDepsTest {

    @Test
    void testSelectDepsWork() {
        assertTrue(true, "Test suite with select() deps should load and run");
    }
}

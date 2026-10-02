package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;
import org.junit.platform.engine.discovery.DiscoverySelectors;
import org.junit.platform.launcher.Launcher;
import org.junit.platform.launcher.LauncherDiscoveryRequest;
import org.junit.platform.launcher.core.LauncherConfig;
import org.junit.platform.launcher.core.LauncherDiscoveryRequestBuilder;
import org.junit.platform.launcher.core.LauncherFactory;

public class CommandLineSummaryTest {

  private static final String FAIL_IF_NO_TESTS_PROPERTY = "JUNIT5_FAIL_IF_NO_TESTS";

  @AfterEach
  public void clearSystemProperty() {
    System.clearProperty(FAIL_IF_NO_TESTS_PROPERTY);
  }

  @Test
  public void getTestCountIsZeroWhenNoTestsAreDiscovered() {
    CommandLineSummary summary = runAgainst(ZeroTestsFixture.class);

    assertEquals(0, summary.getTestCount());
    assertEquals(0, summary.getFailureCount());
  }

  @Test
  public void getTestCountReflectsTheNumberOfDiscoveredTestMethods() {
    CommandLineSummary summary = runAgainst(TwoPassingTestsFixture.class);

    assertEquals(2, summary.getTestCount());
    assertEquals(0, summary.getFailureCount());
  }

  @Test
  public void doesNotFailForNoTestsByDefault() {
    // Mirrors JUnit's own ConsoleLauncher --fail-if-no-tests: disabled unless opted into, so
    // enabling this check is a non-breaking change for existing callers.
    System.clearProperty(FAIL_IF_NO_TESTS_PROPERTY);
    CommandLineSummary summary = runAgainst(ZeroTestsFixture.class);

    assertFalse(ActualRunner.shouldFailForNoTests(summary));
  }

  @Test
  public void failsForNoTestsWhenOptedIn() {
    System.setProperty(FAIL_IF_NO_TESTS_PROPERTY, "true");
    CommandLineSummary summary = runAgainst(ZeroTestsFixture.class);

    assertTrue(ActualRunner.shouldFailForNoTests(summary));
  }

  @Test
  public void doesNotFailWhenOptedInButTestsWereDiscovered() {
    System.setProperty(FAIL_IF_NO_TESTS_PROPERTY, "true");
    CommandLineSummary summary = runAgainst(TwoPassingTestsFixture.class);

    assertFalse(ActualRunner.shouldFailForNoTests(summary));
  }

  private static CommandLineSummary runAgainst(Class<?> testClass) {
    CommandLineSummary summary = new CommandLineSummary();
    LauncherDiscoveryRequest request =
        LauncherDiscoveryRequestBuilder.request()
            .selectors(DiscoverySelectors.selectClass(testClass))
            .build();
    Launcher launcher =
        LauncherFactory.create(LauncherConfig.builder().addTestExecutionListeners(summary).build());
    launcher.execute(request);
    return summary;
  }
}

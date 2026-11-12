package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.PrintWriter;
import java.io.StringWriter;
import org.junit.jupiter.api.Test;
import org.junit.platform.engine.discovery.DiscoverySelectors;
import org.junit.platform.launcher.Launcher;
import org.junit.platform.launcher.LauncherDiscoveryRequest;
import org.junit.platform.launcher.core.LauncherConfig;
import org.junit.platform.launcher.core.LauncherDiscoveryRequestBuilder;
import org.junit.platform.launcher.core.LauncherFactory;

public class CommandLineSummaryTest {

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
    CommandLineSummary summary = runAgainst(ZeroTestsFixture.class);

    assertFalse(ActualRunner.shouldFailForNoTests(summary, /* failIfNoTestsEnabled= */ false));
  }

  @Test
  public void failsForNoTestsWhenOptedIn() {
    CommandLineSummary summary = runAgainst(ZeroTestsFixture.class);

    assertTrue(ActualRunner.shouldFailForNoTests(summary, /* failIfNoTestsEnabled= */ true));
  }

  @Test
  public void doesNotFailWhenOptedInButTestsWereDiscovered() {
    CommandLineSummary summary = runAgainst(TwoPassingTestsFixture.class);

    assertFalse(ActualRunner.shouldFailForNoTests(summary, /* failIfNoTestsEnabled= */ true));
  }

  @Test
  public void hadOutOfMemoryErrorIsFalseWhenTestsPass() {
    CommandLineSummary summary = runAgainst(TwoPassingTestsFixture.class);

    assertFalse(summary.hadOutOfMemoryError());
  }

  @Test
  public void hadOutOfMemoryErrorIsTrueForAnOomFailureWrappedInAnotherThrowable() {
    CommandLineSummary summary = runAgainst(WrappedOomFailureFixture.class);

    assertTrue(summary.hadOutOfMemoryError());
  }

  @Test
  public void hadOutOfMemoryErrorIsTrueForAnOomInACyclicCauseChain() {
    CommandLineSummary summary = runAgainst(CyclicOomFailureFixture.class);

    assertTrue(summary.hadOutOfMemoryError());
  }

  @Test
  public void hadOutOfMemoryErrorIsFalseForAnOrdinaryFailureInACyclicCauseChain() {
    CommandLineSummary summary = runAgainst(CyclicNonOomFailureFixture.class);

    assertFalse(summary.hadOutOfMemoryError());
  }

  @Test
  public void writeToTerminatesOnCyclicCauseChains() {
    CommandLineSummary summary = runAgainst(CyclicNonOomFailureFixture.class);

    StringWriter out = new StringWriter();
    summary.writeTo(new PrintWriter(out));

    assertTrue(out.toString().contains("[CIRCULAR REFERENCE:"));
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

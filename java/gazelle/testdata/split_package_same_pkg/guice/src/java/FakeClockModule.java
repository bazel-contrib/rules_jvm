package com.example.time;

// NOTE: This file is in the SAME Java package as Clock, FakeClock, and Errors,
// but in a DIFFERENT Bazel package.
//
// Since these classes are in the same Java package, no import is needed.
// This tests that Gazelle correctly tracks:
// 1. Same-package type references (Clock, FakeClock)
// 2. Inner class references (ErrorsPresentException is Errors.ErrorsPresentException)

public class FakeClockModule {
    // Uses Clock and FakeClock from the same Java package (no import needed)
    private final Clock clock;
    private final FakeClock fakeClock;

    public FakeClockModule() {
        this.fakeClock = new FakeClock();
        this.clock = fakeClock;
    }

    public Clock getClock() {
        return clock;
    }

    public void validate(Errors errors) {
        // Uses ErrorsPresentException - an inner class of Errors.
        // In Java, inner classes can be referenced by simple name from the same package.
        try {
            errors.throwIfPresent();
        } catch (ErrorsPresentException e) {
            throw new RuntimeException("Validation failed", e);
        }
    }
}

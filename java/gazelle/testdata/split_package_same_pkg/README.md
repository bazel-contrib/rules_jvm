# Split Package - Same Java Package Test

This test demonstrates a known limitation in Gazelle's Java extension when handling
"split packages" where the consumer is in the **same Java package** as the providers.

## Scenario

- `common/src/java/` contains `Clock.java` and `FakeClock.java` in package `com.example.time`
- `guice/src/java/` contains `FakeClockModule.java` also in package `com.example.time`
- `FakeClockModule` uses `Clock` and `FakeClock` without import statements (same package)

## Expected Behavior

The generated BUILD for `guice/src/java/` should include:
```
deps = ["//common/src/java"]
```

## Actual Behavior (Bug)

The generated BUILD has no deps, causing compilation to fail.

## Root Cause

The Java parser (`ClasspathParser.java`) only tracks:
1. Explicit imports (`import com.example.Foo`)
2. Fully qualified type references (`com.example.Foo bar`)

It does NOT track simple type references to same-package classes like `Clock clock;`
because Java doesn't require import statements for same-package classes.

See `ClasspathParser.java:checkFullyQualifiedType()` - when a type reference:
- Is NOT in `currentFileImports` (not explicitly imported)
- Has NO dots (not fully qualified)

Then it's not added to `data.usedTypes`, and Gazelle never learns about the dependency.

## Workaround

Add explicit imports even for same-package classes:
```java
package com.example.time;

import com.example.time.Clock;      // Explicit import for same-package class
import com.example.time.FakeClock;  // Explicit import for same-package class

public class FakeClockModule {
    Clock clock;
    FakeClock fakeClock;
}
```

# Split Package with Private Inner Class Test

This test verifies that Gazelle's Java extension correctly handles split packages
where a file:
1. Uses **private inner classes** defined within the same file
2. References classes from other files in the same package via `.class` literals

## Scenario

- `annotations/src/java/` contains `Uninterruptible.java` (an annotation) with package `com.example.uninterruptibles`
- `lib/src/java/` contains `UninterruptibleListener.java` with the same package, including:
  - A reference to `Uninterruptible.class` (from the annotations package)
  - Private inner classes like `InvokeWithExceptionHandling`
- `app/src/java/` imports `com.example.uninterruptibles.UninterruptibleListener`

## Bugs Fixed

### Bug 1: Private inner class references causing resolution failure
When a file used a private inner class from within the same file, and the
package had multiple providers (split package scenario), resolution would fail.

**Fix**: Track all class names (including private) per file and filter them out
from same-package references at the end of each file's processing.

### Bug 2: Class literals not detected as dependencies
Class literals like `Foo.class` were not being detected as type references,
so dependencies on same-package classes used via `.class` were missing.

**Fix**: Added `visitMemberSelect` to detect class literals and treat them
as same-package type references when the class name is unqualified.

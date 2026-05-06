package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.util.Map;
import java.util.Optional;
import java.util.Set;
import org.junit.jupiter.api.Test;

class TypeNameResolverTest {

  @Test
  void importedSimpleNameResolvesToImportedFqn() {
    assertEquals(
        Optional.of("java.util.UUID"),
        TypeNameResolver.resolve(
            "UUID", Map.of("UUID", "java.util.UUID"), "com.example", Set.of()));
  }

  @Test
  void importedFirstSegmentWinsAndDropsTrailingComponents() {
    // Preserves long-standing ClasspathParser behaviour: `Outer.Inner` where Outer is
    // imported records only the imported name, not Outer.Inner.
    assertEquals(
        Optional.of("com.example.Outer"),
        TypeNameResolver.resolve(
            "Outer.Inner", Map.of("Outer", "com.example.Outer"), "pkg", Set.of()));
  }

  @Test
  void alreadyDottedNameWithoutImportPassesThrough() {
    assertEquals(
        Optional.of("com.example.Foo"),
        TypeNameResolver.resolve("com.example.Foo", Map.of(), "pkg", Set.of()));
  }

  @Test
  void bareNameFallsBackToCurrentPackage() {
    assertEquals(
        Optional.of("com.example.Helper"),
        TypeNameResolver.resolve("Helper", Map.of(), "com.example", Set.of()));
  }

  @Test
  void excludedNameSuppressesSamePackageFallback() {
    assertEquals(
        Optional.empty(),
        TypeNameResolver.resolve("String", Map.of(), "com.example", Set.of("String")));
  }

  @Test
  void nullCurrentPackageDisablesFallback() {
    assertEquals(Optional.empty(), TypeNameResolver.resolve("Helper", Map.of(), null, Set.of()));
  }

  @Test
  void emptyCurrentPackageDisablesFallback() {
    assertEquals(Optional.empty(), TypeNameResolver.resolve("Helper", Map.of(), "", Set.of()));
  }

  @Test
  void emptyTypeNameReturnsEmpty() {
    assertEquals(Optional.empty(), TypeNameResolver.resolve("", Map.of(), "pkg", Set.of()));
  }

  @Test
  void importTakesPrecedenceOverExclusion() {
    // If a name is imported, the import wins regardless of exclusions — exclusions only
    // gate the same-package fallback, not import resolution.
    assertEquals(
        Optional.of("java.util.UUID"),
        TypeNameResolver.resolve("UUID", Map.of("UUID", "java.util.UUID"), "pkg", Set.of("UUID")));
  }
}

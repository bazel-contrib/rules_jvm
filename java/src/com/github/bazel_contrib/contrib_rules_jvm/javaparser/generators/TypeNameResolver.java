package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import javax.annotation.Nullable;
import java.util.Map;
import java.util.Optional;
import java.util.Set;

/**
 * Resolves a (possibly dotted) simple-name string to its fully-qualified form, using imports,
 * already-qualified shape, or a same-package fallback. AST-agnostic so the policy is identical for
 * Java and Kotlin sources — visitors are responsible for extracting the name string and deciding
 * when to call.
 */
final class TypeNameResolver {

  private TypeNameResolver() {}

  /**
   * @param typeName a name such as {@code Foo}, {@code Outer.Inner}, or {@code com.example.Foo}
   * @param imports map from simple name (or alias) to fully-qualified name
   * @param currentPackage the file's own package; null or empty disables the same-package fallback
   * @param excluded names that should NOT trigger the same-package fallback (e.g. java.lang
   *     built-ins, locally-defined classes, type parameters)
   * @return the resolved fully-qualified name, or empty if no policy applies
   */
  static Optional<String> resolve(
      String typeName,
      Map<String, String> imports,
      @Nullable String currentPackage,
      Set<String> excluded) {
    if (typeName == null || typeName.isEmpty()) {
      return Optional.empty();
    }
    int firstDot = typeName.indexOf('.');
    String firstSegment = firstDot == -1 ? typeName : typeName.substring(0, firstDot);
    // An import covering the leading segment wins. Trailing components past the imported name
    // are dropped (preserves existing ClasspathParser behaviour for inputs like "Outer.Inner"
    // where Outer is imported).
    if (imports.containsKey(firstSegment)) {
      return Optional.of(imports.get(firstSegment));
    }
    if (firstDot != -1) {
      // Already FQN-shaped; trust it.
      return Optional.of(typeName);
    }
    if (excluded.contains(typeName)) {
      return Optional.empty();
    }
    if (currentPackage == null || currentPackage.isEmpty()) {
      return Optional.empty();
    }
    return Optional.of(currentPackage + "." + typeName);
  }
}

package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.util.Map;
import java.util.Set;
import java.util.TreeMap;
import java.util.TreeSet;

class ParsedPackageData {
  /** Packages defined. */
  final Set<String> packages = new TreeSet<>();

  /** The fully qualified name of types that are imported. */
  final Set<String> usedTypes = new TreeSet<>();

  /** The name of passages that are imported for wildcards or (in Kotlin) direct function access. */
  final Set<String> usedPackagesWithoutSpecificTypes = new TreeSet<>();

  /** The fully qualified names of types that should be exported by this build rule. */
  final Set<String> exportedTypes = new TreeSet<>();

  /** The short name (no package) of any classes that provide a public static main function. */
  final Set<String> mainClasses = new TreeSet<>();

  /**
   * Simple type names (without package prefix) that are referenced in type contexts but not
   * explicitly imported. These are likely same-package type references that Java doesn't require
   * imports for.
   */
  final Set<String> samePackageTypeReferences = new TreeSet<>();

  /**
   * Fully qualified names of all public classes defined in this package, including inner classes.
   * For example, for a file containing `public class Outer { public static class Inner {} }`, this
   * would contain both "com.example.Outer" and "com.example.Outer.Inner".
   */
  final Set<String> definedClasses = new TreeSet<>();

  /**
   * Maps from fully-qualified class-name to class-names of annotations on that class. Annotations
   * will be fully-qualified where that's known, and not where not known.
   */
  final Map<String, PerClassData> perClassData = new TreeMap<>();

  ParsedPackageData() {}

  void merge(ParsedPackageData other) {
    packages.addAll(other.packages);
    usedTypes.addAll(other.usedTypes);
    usedPackagesWithoutSpecificTypes.addAll(other.usedPackagesWithoutSpecificTypes);
    exportedTypes.addAll(other.exportedTypes);
    mainClasses.addAll(other.mainClasses);
    samePackageTypeReferences.addAll(other.samePackageTypeReferences);
    definedClasses.addAll(other.definedClasses);
    for (Map.Entry<String, PerClassData> classData : other.perClassData.entrySet()) {
      PerClassData existing = perClassData.get(classData.getKey());
      if (existing == null) {
        existing = new PerClassData();
        perClassData.put(classData.getKey(), existing);
      }
      existing.merge(classData.getValue());
    }
  }
}

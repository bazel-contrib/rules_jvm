package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.util.Map;
import java.util.Set;
import java.util.TreeMap;
import java.util.TreeSet;

public class ParsedPackageData {
  /** Packages defined. */
  public final Set<String> packages = new TreeSet<>();

  /** The fully qualified name of types that are imported. */
  public final Set<String> usedTypes = new TreeSet<>();

  /** The name of passages that are imported for wildcards or (in Kotlin) direct function access. */
  public final Set<String> usedPackagesWithoutSpecificTypes = new TreeSet<>();

  /** The fully qualified names of types that should be exported by this build rule. */
  public final Set<String> exportedTypes = new TreeSet<>();

  /**
   * The fully qualified names of top-level {@code internal} declarations in this package. Keyed as
   * an importer would record them (e.g. {@code a.b.foo} for a function), so a depender's imported
   * classes can be matched against them to detect cross-package {@code internal} coupling.
   */
  final Set<String> internalTypes = new TreeSet<>();

  /** The short name (no package) of any classes that provide a public static main function. */
  public final Set<String> mainClasses = new TreeSet<>();

  /**
   * Maps from fully-qualified class-name to class-names of annotations on that class. Annotations
   * will be fully-qualified where that's known, and not where not known.
   */
  public final Map<String, PerClassData> perClassData = new TreeMap<>();

  public ParsedPackageData() {}

  void merge(ParsedPackageData other) {
    packages.addAll(other.packages);
    usedTypes.addAll(other.usedTypes);
    usedPackagesWithoutSpecificTypes.addAll(other.usedPackagesWithoutSpecificTypes);
    exportedTypes.addAll(other.exportedTypes);
    internalTypes.addAll(other.internalTypes);
    mainClasses.addAll(other.mainClasses);
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

package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.util.Map;
import java.util.Set;
import java.util.TreeSet;
import java.util.TreeMap;

class ParsedPackageData {
  final Set<String> packages = new TreeSet<>();
  final Set<String> usedTypes = new TreeSet<>();
  final Set<String> usedPackagesWithoutSpecificTypes = new TreeSet<>();

  final Set<String> exportedTypes = new TreeSet<>();
  final Set<String> mainClasses = new TreeSet<>();

  // Mapping from fully-qualified class-name to class-names of annotations on that class.
  // Annotations will be fully-qualified where that's known, and not where not known.
  final Map<String, PerClassData> perClassData = new TreeMap<>();

  ParsedPackageData() {}
}

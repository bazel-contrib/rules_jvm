package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.util.Map;
import java.util.Objects;
import java.util.SortedMap;
import java.util.SortedSet;
import java.util.TreeMap;
import java.util.TreeSet;

class PerClassData {
  PerClassData() {
    this(new TreeSet<>(), new TreeMap<>(), new TreeMap<>());
  }

  @Override
  public String toString() {
    return "PerClassData{"
        + "annotations="
        + annotations
        + ", perMethodAnnotations="
        + perMethodAnnotations
        + ", perFieldAnnotations="
        + perFieldAnnotations
        + '}';
  }

  PerClassData(
      SortedSet<String> annotations,
      SortedMap<String, SortedSet<String>> perMethodAnnotations,
      SortedMap<String, SortedSet<String>> perFieldAnnotations) {
    this.annotations = annotations;
    this.perMethodAnnotations = perMethodAnnotations;
    this.perFieldAnnotations = perFieldAnnotations;
  }

  final SortedSet<String> annotations;

  final SortedMap<String, SortedSet<String>> perMethodAnnotations;
  final SortedMap<String, SortedSet<String>> perFieldAnnotations;

  public void merge(PerClassData other) {
    annotations.addAll(other.annotations);
    for (Map.Entry<String, SortedSet<String>> methodAndAnnotations :
        other.perMethodAnnotations.entrySet()) {
      SortedSet<String> existing = perMethodAnnotations.get(methodAndAnnotations.getKey());
      if (existing == null) {
        existing = new TreeSet<>();
        perMethodAnnotations.put(methodAndAnnotations.getKey(), existing);
      }
      existing.addAll(methodAndAnnotations.getValue());
    }
    for (Map.Entry<String, SortedSet<String>> fieldAndAnnotations :
        other.perFieldAnnotations.entrySet()) {
      SortedSet<String> existing = perFieldAnnotations.get(fieldAndAnnotations.getKey());
      if (existing == null) {
        existing = new TreeSet<>();
        perFieldAnnotations.put(fieldAndAnnotations.getKey(), existing);
      }
      existing.addAll(fieldAndAnnotations.getValue());
    }
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (o == null || getClass() != o.getClass()) return false;
    PerClassData that = (PerClassData) o;
    return Objects.equals(annotations, that.annotations)
        && Objects.equals(perMethodAnnotations, that.perMethodAnnotations)
        && Objects.equals(perFieldAnnotations, that.perFieldAnnotations);
  }

  @Override
  public int hashCode() {
    return Objects.hash(annotations, perMethodAnnotations, perFieldAnnotations);
  }
}

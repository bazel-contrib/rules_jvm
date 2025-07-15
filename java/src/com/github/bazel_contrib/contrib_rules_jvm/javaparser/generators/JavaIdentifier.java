package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.util.Objects;

public class JavaIdentifier implements Comparable<JavaIdentifier> {

  private final String packageName;
  private final String className;

  /**
   * Copied from the KnowTypeSolvers, this is the bazel dependency string where this package/class
   * will be found: The dependency name will be of the form:
   *
   * <p>- artifact("com.example.library:library") - For a dependency external to the repo and found
   * in the maven cache *
   *
   * <p>- //src/java/com/example/library - For a dependency in the repo the root of the source tree
   * for searching
   *
   * <p>- null - For all dependencies in the default java library or from generated code
   */
  private final String sourceLibrary;

  public JavaIdentifier(String packageName, String className, String sourceLibrary) {
    this.packageName = Objects.requireNonNull(packageName);
    this.className = Objects.requireNonNull(className);
    this.sourceLibrary = sourceLibrary;
  }

  public String getSourceLibrary() {
    return sourceLibrary;
  }

  @Override
  public int compareTo(JavaIdentifier that) {
    return (packageName + "." + className).compareTo(that.packageName + "." + that.className);
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) {
      return true;
    }

    if (!(o instanceof JavaIdentifier)) {
      return false;
    }

    JavaIdentifier that = (JavaIdentifier) o;
    return Objects.equals(this.packageName, that.packageName)
        && Objects.equals(this.className, that.className);
  }

  @Override
  public int hashCode() {
    return Objects.hash(packageName, className);
  }

  @Override
  public String toString() {
    return packageName + "." + className + " -> " + sourceLibrary;
  }
}

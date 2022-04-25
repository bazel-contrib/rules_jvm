package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.io.IOException;
import java.nio.file.FileVisitResult;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.SimpleFileVisitor;
import java.nio.file.attribute.BasicFileAttributes;

public class BuildFileVisitor extends SimpleFileVisitor<Path> {

  private final PackageDependencies dependencies;

  public BuildFileVisitor(PackageDependencies dependencies) {
    this.dependencies = dependencies;
  }

  @Override
  public FileVisitResult preVisitDirectory(Path dir, BasicFileAttributes attrs) throws IOException {
    // if this is the base directory for the package, look through this one
    if (dir.equals(dependencies.getPackagePath().getParent())) {
      return FileVisitResult.CONTINUE;
      // If this is a subdir with a build file, skip the directory, there is another package
      // depenency waiting
    } else if (Files.exists(dir.resolve(dependencies.getPackagePath().getFileName()))) {
      return FileVisitResult.SKIP_SUBTREE;
      // else read this directory
    } else {
      return FileVisitResult.CONTINUE;
    }
  }

  @Override
  public FileVisitResult visitFile(Path file, BasicFileAttributes attrs) throws IOException {
    if (file.toString().endsWith(".java")) {
      dependencies.resolveTypesForClass(file);
    }
    return FileVisitResult.CONTINUE;
  }
}

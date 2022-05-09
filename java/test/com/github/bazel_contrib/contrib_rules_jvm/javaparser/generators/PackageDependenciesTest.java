package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static org.junit.jupiter.api.Assertions.assertEquals;

import com.github.bazel_contrib.contrib_rules_jvm.javaparser.file.BuildFile;
import edu.umd.cs.findbugs.annotations.SuppressFBWarnings;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class PackageDependenciesTest {
  private final List<PackageDependencies> packages = new ArrayList<>();
  private final KnownTypeSolvers solvers = new KnownTypeSolvers(packages);

  private Path root;
  private Path sources;
  private Path packagePath;
  private Path rootPath;

  @BeforeEach
  @SuppressFBWarnings("DMI_HARDCODED_ABSOLUTE_FILENAME")
  void createPaths() {
    root = Paths.get("/home", "workspace");
    sources = Paths.get("src", "main");
    packagePath = Paths.get("com", "gazelle", "java", "javaparser", "BUILD.bazel");
    rootPath = root.resolve(sources).resolve(packagePath);
  }

  @Test
  void targetNameTest() {
    String bazelTarget = String.format("//%s", root.relativize(rootPath).getParent());

    PackageDependencies deps =
        new PackageDependencies(new BuildFile(rootPath, bazelTarget, null), solvers);

    assertEquals(deps.getTargetName(), "javaparser");
  }

  @Test
  void packagePathTest() {
    String bazelTarget = String.format("//%s", root.relativize(rootPath).getParent());

    PackageDependencies deps =
        new PackageDependencies(new BuildFile(rootPath, bazelTarget, null), solvers);

    assertEquals(
        "/home/workspace/src/main/com/gazelle/java/javaparser/BUILD.bazel",
        deps.getPackagePath().toString());
  }

  @Test
  void packageNameTest() {
    String bazelTarget = String.format("//%s", root.relativize(rootPath).getParent());

    PackageDependencies deps =
        new PackageDependencies(new BuildFile(rootPath, bazelTarget, null), solvers);
    assertEquals("//src/main/com/gazelle/java/javaparser", deps.getBuildFile().getBazelTarget());
  }
}

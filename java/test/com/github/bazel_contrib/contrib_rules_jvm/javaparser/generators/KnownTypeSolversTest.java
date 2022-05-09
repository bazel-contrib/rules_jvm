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

public class KnownTypeSolversTest {
  private KnownTypeSolvers solvers;
  private final List<PackageDependencies> packages = new ArrayList<>();
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

    solvers = new KnownTypeSolvers(packages);

    String bazelTarget = String.format("//%s", root.relativize(rootPath).getParent());

    packages.add(new PackageDependencies(new BuildFile(rootPath, bazelTarget, null), solvers));
    Path deeperPath =
        Paths.get("com", "gazelle", "java", "javaparser", "annotations", "BUILD.bazel");
    Path deeperRoot = root.resolve(sources).resolve(deeperPath);
    bazelTarget = String.format("//%s", root.relativize(deeperRoot).getParent());
    packages.add(new PackageDependencies(new BuildFile(deeperRoot, bazelTarget, null), solvers));
  }

  @Test
  void validateGetSourceBazelTarget() {
    String target =
        solvers.getSourceBazelTarget("//src/main", List.of("com", "gazelle", "java", "javaparser"));
    assertEquals("//src/main/com/gazelle/java/javaparser", target);
  }

  @Test
  void validateOtherGetSourceBazelTarget() {
    String target =
        solvers.getSourceBazelTarget(
            "//src/main", List.of("com", "gazelle", "java", "javaparser", "lava"));
    assertEquals("//src/main/com/gazelle/java/javaparser", target);
  }

  @Test
  void validateDeeperSourceBazelTarget() {
    String target =
        solvers.getSourceBazelTarget(
            "//src/main", List.of("com", "gazelle", "java", "javaparser", "annotations"));
    assertEquals("//src/main/com/gazelle/java/javaparser/annotations", target);
  }
}

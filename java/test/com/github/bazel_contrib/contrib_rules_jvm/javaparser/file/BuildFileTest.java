package com.github.bazel_contrib.contrib_rules_jvm.javaparser.file;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.nio.file.Paths;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

public class BuildFileTest {

  Path root;
  Path source;
  Path packagePath;
  Path build;
  Path rootPath;

  @BeforeEach
  void createPaths() {
    root = Paths.get("home", "workspace");
    source = Paths.get("src", "main");
    packagePath = Paths.get("com", "gazelle", "java", "generators");
    build = Paths.get("BUILD.bazel");
    rootPath = root.resolve(source).resolve(packagePath).resolve(build);
  }

  @Test
  void testTarget() {
    BuildFile buildFile = new BuildFile(rootPath, "//" + source.resolve(packagePath), null);
    assertEquals(buildFile.getTargetName(), "generators");
  }

  @Test
  void testBazelTarget() {
    BuildFile buildFile = new BuildFile(rootPath, "//" + source.resolve(packagePath), null);
    assertEquals("//src/main/com/gazelle/java/generators", buildFile.getBazelTarget());
  }

  @Test
  void testReplace() throws IOException {
    String content;
    try (InputStream in = this.getClass().getResourceAsStream("build.source.txt")) {
      content = new String(in.readAllBytes(), StandardCharsets.UTF_8);
    }
    BuildFile buildFile = new BuildFile(rootPath, "//" + source.resolve(packagePath), content);

    String deps =
        "generators_deps = [\n"
            + "  \"//src/main/com/gazelle/java\",\n"
            + "   artifact(\"com.gazelle.java:package\"),\n"
            + "]\n";
    String update = buildFile.replaceTarget("_deps", deps);

    Assertions.assertTrue(update.contains(deps));
  }
}

package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.nio.file.FileSystems;
import java.nio.file.FileVisitOption;
import java.nio.file.FileSystem;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.SortedSet;
import java.util.TreeMap;
import java.util.TreeSet;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class KtParserTest {
  private static final Logger logger = LoggerFactory.getLogger(KtParserTest.class);

  private static Path workspace;
  private static Path directory;

  private KtParser parser;

  @BeforeAll
  public static void setup() throws IOException, URISyntaxException {
    URI workspaceUri = KtParserTest.class.getClassLoader().getResource("workspace").toURI();
    try (FileSystem fileSystem = FileSystems.newFileSystem(workspaceUri, new HashMap<>())) {
      // The IntelliJ file manager doesn't support reading resources from a jar, so we need to
      // extract them into a temporary directory before accessing them from the KtParser.
      Path workspaceResourcePath = Paths.get(workspaceUri);
      Path directoryResourcePath = workspaceResourcePath.resolve("com/gazelle/kotlin/javaparser/generators");

      workspace = Files.createTempDirectory("workspace");
      directory = workspace.resolve("com/gazelle/kotlin/javaparser/generators");
      directory = Files.createDirectories(directory);

      Files.walk(directoryResourcePath).forEach(file -> {
        if (Files.isDirectory(file)) {
            return;
        }
        try {
            byte[] bytes = Files.readAllBytes(file);
            Files.write(directory.resolve(file.getFileName().toString()), bytes);
        } catch (Exception e) {
          logger.error("Error copying file " + file.toString(), e);
        }
      });
    }
  }

  @BeforeEach
  public void setupPerTest() {
    parser = new KtParser();
  }

  @Test
  public void topLevelMainFunction() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("Main.kt"));

    assertEquals(Set.of("workspace.com.gazelle.kotlin.javaparser.generators"), data.packages);
    assertEquals(Set.of("MainKt"), data.mainClasses);
    assertEquals(Set.of("workspace.com.gazelle.kotlin.javaparser.generators.MainKt"), data.perClassData.keySet());
  }

  @Test
  public void mainInClass() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("MainInClass.kt"));

    assertEquals(Set.of("workspace.com.gazelle.kotlin.javaparser.generators"), data.packages);
    assertEquals(Set.of("MainInClass"), data.mainClasses);
    assertEquals(Set.of("workspace.com.gazelle.kotlin.javaparser.generators.MainInClass", "workspace.com.gazelle.kotlin.javaparser.generators.MainInClass.Companion"), data.perClassData.keySet());
  }

  @Test
  public void mainOnCompanion() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("MainOnCompanion.kt"));

    assertEquals(Set.of("workspace.com.gazelle.kotlin.javaparser.generators"), data.packages);
    assertEquals(Set.of("MainOnCompanion.Companion"), data.mainClasses);
  }

  @Test
  public void exportingClassTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("ExportingClass.kt"));

    assertEquals(Set.of("example.external.InternalReturn", "example.external.ProtectedReturn", "example.external.PublicReturn", "example.external.ParameterizedReturn"), data.exportedTypes);
  }

  @Test
  public void helloTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("Hello.kt"));

    assertEquals(Set.of("workspace.com.gazelle.kotlin.javaparser.generators"), data.packages);
    assertEquals(Set.of("com.gazelle.java.javaparser.generators.DeleteBookRequest", "com.gazelle.java.javaparser.generators.HelloProto", "com.google.common.primitives.Ints"), data.usedTypes);
    assertEquals(Set.of(), data.usedPackagesWithoutSpecificTypes);
    assertEquals(Set.of(), data.exportedTypes);
    assertEquals(Set.of(), data.mainClasses);
    assertEquals(Set.of("workspace.com.gazelle.kotlin.javaparser.generators.Hello"), data.perClassData.keySet());
  }

  @Test
  public void staticImportsTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("StaticImports.kt"));

    assertEquals(Set.of("com.gazelle.java.javaparser.ClasspathParser"), data.usedTypes);
    assertEquals(Set.of("com.gazelle.kotlin.constantpackage", "com.gazelle.kotlin.constantpackage2", "com.gazelle.kotlin.functionpackage"), data.usedPackagesWithoutSpecificTypes);
  }

  @Test
  public void wildcardsTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("Wildcards.kt"));

    assertEquals(Set.of("org.junit.jupiter.api"), data.usedPackagesWithoutSpecificTypes);
    assertEquals(Set.of("org.junit.jupiter.api.Assertions"), data.usedTypes);
  }

  private List<Path> getPathsWithNames(String... names) throws IOException {
    Set<String> namesSet = Set.of(names);
    return Files.walk(directory).filter(file -> namesSet.contains(file.getFileName().toString())).toList();
  }
}

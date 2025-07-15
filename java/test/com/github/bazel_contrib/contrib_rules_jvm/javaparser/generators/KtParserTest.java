package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.nio.file.FileSystem;
import java.nio.file.FileSystems;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.List;
import java.util.Set;
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
    try (@SuppressWarnings("unused")
        FileSystem fileSystem = FileSystems.newFileSystem(workspaceUri, new HashMap<>())) {
      // The IntelliJ file manager doesn't support reading resources from a jar, so we need to
      // extract them into a temporary directory before accessing them from the KtParser.
      Path workspaceResourcePath = Paths.get(workspaceUri);
      Path directoryResourcePath =
          workspaceResourcePath.resolve("com/gazelle/kotlin/javaparser/generators");

      workspace = Files.createTempDirectory("workspace");
      directory = workspace.resolve("com/gazelle/kotlin/javaparser/generators");
      directory = Files.createDirectories(directory);

      Files.walk(directoryResourcePath)
          .forEach(
              file -> {
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
    assertEquals(
        Set.of("workspace.com.gazelle.kotlin.javaparser.generators.MainKt"),
        data.perClassData.keySet());
  }

  @Test
  public void mainInClass() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("MainInClass.kt"));

    assertEquals(Set.of("workspace.com.gazelle.kotlin.javaparser.generators"), data.packages);
    assertEquals(Set.of("MainInClass"), data.mainClasses);
    assertEquals(
        Set.of(
            "workspace.com.gazelle.kotlin.javaparser.generators.MainInClass",
            "workspace.com.gazelle.kotlin.javaparser.generators.MainInClass.Companion"),
        data.perClassData.keySet());
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

    assertEquals(
        Set.of(
            "example.external.FinalProperty",
            "example.external.VarProperty",
            "example.external.InternalReturn",
            "example.external.ProtectedReturn",
            "example.external.PublicReturn",
            "example.external.ParameterizedReturn"),
        data.exportedTypes);
  }

  @Test
  public void privateExportingClassTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("PrivateExportingClass.kt"));

    assertEquals(Set.of(), data.exportedTypes);
  }

  @Test
  public void helloTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("Hello.kt"));

    assertEquals(Set.of("workspace.com.gazelle.kotlin.javaparser.generators"), data.packages);
    assertEquals(
        Set.of(
            "com.gazelle.java.javaparser.generators.DeleteBookRequest",
            "com.gazelle.java.javaparser.generators.HelloProto",
            "com.google.common.primitives.Ints"),
        data.usedTypes);
    assertEquals(Set.of(), data.usedPackagesWithoutSpecificTypes);
    assertEquals(Set.of(), data.exportedTypes);
    assertEquals(Set.of(), data.mainClasses);
    assertEquals(
        Set.of("workspace.com.gazelle.kotlin.javaparser.generators.Hello"),
        data.perClassData.keySet());
  }

  @Test
  public void constantTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("Constant.kt"));

    assertEquals(
        Set.of("workspace.com.gazelle.kotlin.javaparser.generators.ConstantKt"),
        data.perClassData.keySet());
  }

  // @Test
  // public void fullyQualifiedClassAndFunctionUse() throws IOException {
  //   ParsedPackageData data = parser.parseClasses(getPathsWithNames("FullyQualifieds.kt"));
  //   assertEquals(
  //     Set.of("com.example"),
  //     data.usedPackagesWithoutSpecificTypes);
  //   assertEquals(
  //       Set.of(
  //           "workspace.com.gazelle.java.javaparser.generators.DeleteBookRequest",
  //           "workspace.com.gazelle.java.javaparser.generators.DeleteBookResponse",
  //           "workspace.com.gazelle.java.javaparser.utils.Printer",
  //           "workspace.com.gazelle.java.javaparser.factories.Factory",
  //           "java.util.ArrayList",
  //           "com.example.PrivateArg"),
  //       data.usedTypes);
  // }

  @Test
  public void staticImportsTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("StaticImports.kt"));

    assertEquals(Set.of("com.gazelle.java.javaparser.ClasspathParser"), data.usedTypes);
    assertEquals(
        Set.of(
            "com.gazelle.kotlin.constantpackage",
            "com.gazelle.kotlin.constantpackage2",
            "com.gazelle.kotlin.functionpackage"),
        data.usedPackagesWithoutSpecificTypes);
  }

  @Test
  public void wildcardsTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("Wildcards.kt"));

    assertEquals(Set.of("org.junit.jupiter.api"), data.usedPackagesWithoutSpecificTypes);
    assertEquals(Set.of("org.junit.jupiter.api.Assertions"), data.usedTypes);
  }

  @Test
  public void detectsInlineFunction() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("InlineFunction.kt"));

    // Verify inline function is detected
    assertNotNull(data.implicitDeps, "inlineFunctions map should not be null");
    assertEquals(
        Set.of("com.example.Helper", "com.google.gson.Gson"),
        data.implicitDeps,
        "Should detect the inline function processData: " + data.implicitDeps);
  }

  @Test
  public void detectsMultipleInlineFunctions() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("MultipleInlines.kt"));

    assertNotNull(data.implicitDeps, "inlineFunctions map should not be null");

    // Should detect all three inline functions
    assertEquals(
        Set.of("com.example.utils.StringUtils", "java.util.ArrayList"),
        data.implicitDeps,
        String.valueOf(data.implicitDeps));
  }

  @Test
  public void detectsGsonAndArrayListInInlineFunction() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("InlineWithGson.kt"));

    assertNotNull(data.implicitDeps, "inlineFunctions map should not be null");

    // Should detect the processData inline function
    assertEquals(
        Set.of("com.google.gson.Gson", "java.util.ArrayList"),
        data.implicitDeps,
        String.valueOf(data.implicitDeps));
  }

  private List<Path> getPathsWithNames(String... names) throws IOException {
    Set<String> namesSet = Set.of(names);
    return Files.walk(directory)
        .filter(file -> namesSet.contains(file.getFileName().toString()))
        .toList();
  }

  @Test
  public void detectsSimpleExtensionFunctions() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("SimpleExtensions.kt"));

    assertEquals(
        Set.of("com.example.Helper", "com.google.gson.Gson", "com.google.gson.JsonArray"),
        data.implicitDeps);
  }

  @Test
  public void detectsExtensionOperators() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("ExtensionOperators.kt"));

    assertEquals(
        Set.of("com.example.MathUtils", "com.google.gson.JsonArray", "com.google.gson.JsonObject"),
        data.implicitDeps);
  }

  @Test
  public void testAstEnhancementsAreActive() throws IOException {
    // This test verifies that our AST-based enhancements are active and working
    // by checking that the enhanced visitor methods are being called

    // Test with inline functions
    ParsedPackageData inlineData = parser.parseClasses(getPathsWithNames("InlineWithGson.kt"));
    assertEquals(
        Set.of("com.google.gson.Gson", "java.util.ArrayList"),
        inlineData.implicitDeps,
        "inlineFunctions map should not be null");

    // Test with extension functions
    ParsedPackageData extensionData = parser.parseClasses(getPathsWithNames("SimpleExtensions.kt"));
    assertEquals(
        Set.of("com.example.Helper", "com.google.gson.Gson", "com.google.gson.JsonArray"),
        extensionData.implicitDeps,
        "extensionFunctions should not be null");

    // Test with extension operators
    ParsedPackageData operatorData =
        parser.parseClasses(getPathsWithNames("ExtensionOperators.kt"));
    assertEquals(
        Set.of("com.example.MathUtils", "com.google.gson.JsonArray", "com.google.gson.JsonObject"),
        operatorData.implicitDeps,
        "extensionFunctions should not be null for operators");
  }

  @Test
  public void detectsDestructuringWithCustomComponentFunctions() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("DestructuringWithDeps.kt"));

    assertNotNull(data.implicitDeps, "implicitDeps should not be null");

    // Log what we found for debugging
    logger.info("=== Destructuring Detection Test ===");
    logger.info("Implicit deps found: " + data.implicitDeps);

    // Should detect dependencies from componentN() functions
    assertTrue(
        data.implicitDeps.contains("com.google.gson.Gson")
            || data.implicitDeps.contains("com.google.code.gson.Gson"),
        "Should detect Gson dependency from component1() function. Found: " + data.implicitDeps);

    assertTrue(
        data.implicitDeps.contains("com.google.common.base.Strings"),
        "Should detect Guava Strings dependency from component2() function. Found: "
            + data.implicitDeps);

    // Verify that we have at least some implicit dependencies
    assertTrue(
        data.implicitDeps.size() > 0,
        "Should have detected some implicit dependencies from componentN() functions");

    logger.info("Destructuring detection working correctly!");
  }
}

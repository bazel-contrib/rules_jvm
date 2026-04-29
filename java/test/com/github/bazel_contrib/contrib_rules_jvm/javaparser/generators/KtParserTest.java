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
import java.util.stream.Collectors;
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

  @Test
  public void testFqnExpressionsDetected() throws IOException {
    // Gap: visitDotQualifiedExpression doesn't reconstruct FQN class references.
    // FQN constructor calls (com.example.Foo()) and FQN static calls
    // (com.example.Foo.method()) bypass imports entirely and are invisible.
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("FullyQualifieds.kt"));

    // FQN constructor call: workspace.com.gazelle.java.javaparser.generators.DeleteBookRequest()
    assertTrue(
        data.usedTypes.contains(
            "workspace.com.gazelle.java.javaparser.generators.DeleteBookRequest"),
        "Should detect FQN constructor call. Found: " + data.usedTypes);

    // FQN class reference: workspace.com.gazelle.java.javaparser.utils.Printer.print()
    assertTrue(
        data.usedTypes.contains("workspace.com.gazelle.java.javaparser.utils.Printer"),
        "Should detect FQN class in static method call. Found: " + data.usedTypes);

    // FQN class reference: workspace.com.gazelle.java.javaparser.factories.Factory.create()
    assertTrue(
        data.usedTypes.contains("workspace.com.gazelle.java.javaparser.factories.Factory"),
        "Should detect FQN class in static method call. Found: " + data.usedTypes);

    // FQN constructor call: java.util.ArrayList<String>()
    assertTrue(
        data.usedTypes.contains("java.util.ArrayList"),
        "Should detect FQN constructor call for ArrayList. Found: " + data.usedTypes);
  }

  @Test
  public void testFqnTypePositionsDetected() throws IOException {
    // Gap: tryGetFullyQualifiedName calls KtUserType.getReferencedName() which returns
    // only the last segment (e.g. "InputData"), never the full "com.example.types.InputData".
    // The qualifier chain is never walked, so FQN types in all type positions are invisible.
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("FqnTypePositions.kt"));

    // FQN function parameter type
    assertTrue(
        data.usedTypes.contains("com.example.types.InputData"),
        "Should detect FQN parameter type. Found: " + data.usedTypes);

    // FQN function return type
    assertTrue(
        data.usedTypes.contains("com.example.types.OutputData"),
        "Should detect FQN return type. Found: " + data.usedTypes);

    // FQN property type
    assertTrue(
        data.usedTypes.contains("com.example.config.AppConfig"),
        "Should detect FQN property type. Found: " + data.usedTypes);

    // FQN in is-check
    assertTrue(
        data.usedTypes.contains("com.example.types.Marker"),
        "Should detect FQN in is-check. Found: " + data.usedTypes);

    // FQN in as-cast
    assertTrue(
        data.usedTypes.contains("com.example.types.Castable"),
        "Should detect FQN in as-cast. Found: " + data.usedTypes);
  }

  @Test
  public void testFqnAnnotationsDetected() throws IOException {
    // Gap: FQN annotations (@com.example.MyAnnotation) bypass import handling entirely.
    // In Kotlin PSI, annotation types are KtUserType nodes with qualifier chains,
    // but no annotation-specific handling routes them through FQN detection.
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("FqnAnnotations.kt"));

    assertTrue(
        data.usedTypes.contains("com.example.annotations.ClassAnnotation"),
        "Should detect FQN class annotation. Found: " + data.usedTypes);

    assertTrue(
        data.usedTypes.contains("com.example.annotations.FieldAnnotation"),
        "Should detect FQN field annotation. Found: " + data.usedTypes);

    assertTrue(
        data.usedTypes.contains("com.example.annotations.MethodAnnotation"),
        "Should detect FQN method annotation. Found: " + data.usedTypes);
  }

  @Test
  public void testSamePackageInAllTypePositions() throws IOException {
    // Mirrors ClasspathParserTest.testSamePackageAllTypePositions: same-package
    // class names used as supertype, field type, parameter type, return type,
    // is-check, and as-cast should all flow through the resolver and end up in
    // usedTypes with the file's package prefixed. Built-in names like Boolean
    // and Any must not be added (kotlin.* skip list).
    ParsedPackageData data =
        parser.parseClasses(getPathsWithNames("SamePackageAllPositions.kt"));

    assertEquals(
        Set.of(
            "workspace.com.gazelle.kotlin.javaparser.generators.SomeSuperType",
            "workspace.com.gazelle.kotlin.javaparser.generators.SomeFieldType",
            "workspace.com.gazelle.kotlin.javaparser.generators.SomeParamType",
            "workspace.com.gazelle.kotlin.javaparser.generators.SomeReturnType",
            "workspace.com.gazelle.kotlin.javaparser.generators.SomeMarker",
            "workspace.com.gazelle.kotlin.javaparser.generators.SomeCastTarget"),
        data.usedTypes);
  }

  @Test
  public void testBareClassMethodReceiver() throws IOException {
    // Gap: KtParser relies entirely on imports for usedTypes. A same-package class
    // used as a bare method receiver (SamePackageHelper.create()) has no import,
    // and visitSimpleNameExpression never adds unresolved class names to usedTypes.
    // This matters for split packages where the class is in a different Bazel target.
    // ClasspathParser handles this via checkFullyQualifiedType's same-package fallback.
    ParsedPackageData data =
        parser.parseClasses(getPathsWithNames("BareClassMethodReceiver.kt"));

    assertTrue(
        data.usedTypes.contains(
            "workspace.com.gazelle.kotlin.javaparser.generators.SamePackageHelper"),
        "Should detect same-package bare class method receiver. Found: " + data.usedTypes);
  }

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

    // Verify inline function dependencies are in exported types
    assertNotNull(data.exportedTypes, "exportedTypes should not be null");
    assertTrue(
        data.exportedTypes.contains("com.example.Helper"),
        "Should detect Helper dependency from inline function: " + data.exportedTypes);
    assertTrue(
        data.exportedTypes.contains("com.google.gson.Gson"),
        "Should detect Gson dependency from inline function: " + data.exportedTypes);
  }

  @Test
  public void detectsMultipleInlineFunctions() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("MultipleInlines.kt"));

    assertNotNull(data.exportedTypes, "exportedTypes should not be null");

    // Should detect all three inline functions' dependencies
    assertTrue(
        data.exportedTypes.contains("com.example.utils.StringUtils"),
        "Should detect StringUtils from inline functions: " + data.exportedTypes);
    assertTrue(
        data.exportedTypes.contains("java.util.ArrayList"),
        "Should detect ArrayList from inline functions: " + data.exportedTypes);
  }

  @Test
  public void detectsGsonAndArrayListInInlineFunction() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("InlineWithGson.kt"));

    assertNotNull(data.exportedTypes, "exportedTypes should not be null");

    // Should detect the processData inline function's dependencies
    assertTrue(
        data.exportedTypes.contains("com.google.gson.Gson"),
        "Should detect Gson from inline function: " + data.exportedTypes);
    assertTrue(
        data.exportedTypes.contains("java.util.ArrayList"),
        "Should detect ArrayList from inline function: " + data.exportedTypes);
  }

  private List<Path> getPathsWithNames(String... names) throws IOException {
    Set<String> namesSet = Set.of(names);
    return Files.walk(directory)
        .filter(file -> namesSet.contains(file.getFileName().toString()))
        .collect(Collectors.toUnmodifiableList());
  }

  @Test
  public void detectsSimpleExtensionFunctions() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("SimpleExtensions.kt"));

    // Extension function dependencies should be in exported types
    assertTrue(
        data.exportedTypes.contains("com.example.Helper"),
        "Should detect Helper from extension function: " + data.exportedTypes);
    assertTrue(
        data.exportedTypes.contains("com.google.gson.Gson"),
        "Should detect Gson from extension function: " + data.exportedTypes);
    assertTrue(
        data.exportedTypes.contains("com.google.gson.JsonArray"),
        "Should detect JsonArray from extension function: " + data.exportedTypes);
  }

  @Test
  public void detectsExtensionOperators() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("ExtensionOperators.kt"));

    // Extension operator dependencies should be in exported types
    assertTrue(
        data.exportedTypes.contains("com.example.MathUtils"),
        "Should detect MathUtils from extension operator: " + data.exportedTypes);
    assertTrue(
        data.exportedTypes.contains("com.google.gson.JsonArray"),
        "Should detect JsonArray from extension operator: " + data.exportedTypes);
    assertTrue(
        data.exportedTypes.contains("com.google.gson.JsonObject"),
        "Should detect JsonObject from extension operator: " + data.exportedTypes);
  }

  @Test
  public void testAstEnhancementsAreActive() throws IOException {
    // This test verifies that our AST-based enhancements are active and working
    // by checking that the enhanced visitor methods are being called

    // Test with inline functions - dependencies should be in exportedTypes
    ParsedPackageData inlineData = parser.parseClasses(getPathsWithNames("InlineWithGson.kt"));
    assertTrue(
        inlineData.exportedTypes.contains("com.google.gson.Gson"),
        "Should detect Gson in exportedTypes: " + inlineData.exportedTypes);
    assertTrue(
        inlineData.exportedTypes.contains("java.util.ArrayList"),
        "Should detect ArrayList in exportedTypes: " + inlineData.exportedTypes);

    // Test with extension functions - dependencies should be in exportedTypes
    ParsedPackageData extensionData = parser.parseClasses(getPathsWithNames("SimpleExtensions.kt"));
    assertTrue(
        extensionData.exportedTypes.contains("com.example.Helper"),
        "Should detect Helper in exportedTypes: " + extensionData.exportedTypes);
    assertTrue(
        extensionData.exportedTypes.contains("com.google.gson.Gson"),
        "Should detect Gson in exportedTypes: " + extensionData.exportedTypes);
    assertTrue(
        extensionData.exportedTypes.contains("com.google.gson.JsonArray"),
        "Should detect JsonArray in exportedTypes: " + extensionData.exportedTypes);

    // Test with extension operators - dependencies should be in exportedTypes
    ParsedPackageData operatorData =
        parser.parseClasses(getPathsWithNames("ExtensionOperators.kt"));
    assertTrue(
        operatorData.exportedTypes.contains("com.example.MathUtils"),
        "Should detect MathUtils in exportedTypes: " + operatorData.exportedTypes);
    assertTrue(
        operatorData.exportedTypes.contains("com.google.gson.JsonArray"),
        "Should detect JsonArray in exportedTypes: " + operatorData.exportedTypes);
    assertTrue(
        operatorData.exportedTypes.contains("com.google.gson.JsonObject"),
        "Should detect JsonObject in exportedTypes: " + operatorData.exportedTypes);
  }

  @Test
  public void detectsDestructuringWithCustomComponentFunctions() throws IOException {
    ParsedPackageData data = parser.parseClasses(getPathsWithNames("DestructuringWithDeps.kt"));

    assertNotNull(data.exportedTypes, "exportedTypes should not be null");

    // Log what we found for debugging
    logger.info("=== Destructuring Detection Test ===");
    logger.info("Exported types found: " + data.exportedTypes);

    // Should detect dependencies from componentN() functions in exportedTypes
    assertTrue(
        data.exportedTypes.contains("com.google.gson.Gson")
            || data.exportedTypes.contains("com.google.code.gson.Gson"),
        "Should detect Gson dependency from component1() function. Found: " + data.exportedTypes);

    assertTrue(
        data.exportedTypes.contains("com.google.common.base.Strings"),
        "Should detect Guava Strings dependency from component2() function. Found: "
            + data.exportedTypes);

    // Verify that we have at least some exported types from componentN() functions
    assertTrue(
        data.exportedTypes.size() > 0,
        "Should have detected some exported types from componentN() functions");

    logger.info("Destructuring detection working correctly!");
  }
}

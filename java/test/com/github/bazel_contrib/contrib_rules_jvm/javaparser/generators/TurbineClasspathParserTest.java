package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static org.junit.jupiter.api.Assertions.assertEquals;

import edu.umd.cs.findbugs.annotations.SuppressFBWarnings;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.nio.file.FileSystems;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.SortedSet;
import java.util.TreeMap;
import java.util.TreeSet;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.condition.EnabledIf;
import org.junit.jupiter.api.io.TempDir;

@EnabledIf("isJava21OrHigher")
public class TurbineClasspathParserTest {

  static boolean isJava21OrHigher() {
    return Runtime.version().feature() >= 21;
  }

  private static Path workspace;
  private static Path directory;

  private TurbineClasspathParser parser;

  @BeforeAll
  @SuppressFBWarnings("RCN_REDUNDANT_NULLCHECK_WOULD_HAVE_BEEN_A_NPE")
  public static void setup() throws IOException, URISyntaxException {
    URI uri = TurbineClasspathParserTest.class.getClassLoader().getResource("workspace").toURI();
    Map<String, String> env = new HashMap<>();
    env.put("create", "true");
    FileSystems.newFileSystem(uri, env);
    workspace = Paths.get(uri);
    directory = workspace.resolve("com/gazelle/java/javaparser/generators");
  }

  @BeforeEach
  public void setupPerTest() {
    parser = new TurbineClasspathParser();
  }

  @Test
  public void pathTest() {
    assertEquals("/workspace", workspace.toString());
    assertEquals("/workspace/com/gazelle/java/javaparser/generators", directory.toString());
  }

  @Test
  public void simpleTest() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("Main.java"));
    assertEquals(Set.of("workspace.com.gazelle.java.javaparser.generators"), data.packages);
    assertEquals(Set.of("Main"), data.mainClasses);
  }

  @Test
  public void verifyPackages() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("Main.java"));
    assertEquals(Set.of("workspace.com.gazelle.java.javaparser.generators"), data.packages);
  }

  @Test
  public void verifyNoPackageWithAnnotation() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("NoPackage.java"));
    assertEquals(Set.of(), data.packages);
  }

  @Test
  public void verifyMainClasses() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("Main.java"));
    assertEquals(Set.of("Main"), data.mainClasses);
  }

  @Test
  public void verifyNoMainClasses() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("ClasspathParser.java"));
    assertEquals(0, data.mainClasses.size());
  }

  @Test
  public void verifyPackagesUnique() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(
            directory, List.of("Main.java", "PackageParser.java", "ClasspathParser.java"));
    assertEquals(Set.of("workspace.com.gazelle.java.javaparser.generators"), data.packages);
  }

  @Test
  public void verifyImportsOnParse() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("Hello.java"));
    assertEquals(
        Set.of(
            "com.google.common.primitives.Ints",
            "workspace.com.gazelle.java.javaparser.generators.DeleteBookRequest",
            "workspace.com.gazelle.java.javaparser.generators.HelloProto"),
        data.usedTypes);
  }

  @Test
  public void testWildcardImportOverlap() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("Wildcards.java"));
    assertEquals(Set.of("org.junit.jupiter.api.Assertions"), data.usedTypes);
    assertEquals(Set.of("org.junit.jupiter.api"), data.usedPackagesWithoutSpecificTypes);
  }

  @Test
  public void testFullyQualifiedClassUseNotViaImport() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("PackageParser.java"));
    assertEquals(
        Set.of(
            "com.gazelle.java.ArrayParse",
            "com.gazelle.java.ClasspathParser",
            "com.gazelle.java.OtherClasspathParse"),
        data.usedTypes);
  }

  @Test
  public void testStaticImport() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("StaticImports.java"));
    assertEquals(Set.of("com.gazelle.java.javaparser.ClasspathParser"), data.usedTypes);
  }

  @Test
  public void testWildcardImport() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("WildcardImport.java"));
    assertEquals(Set.of(), data.usedTypes);
    assertEquals(Set.of("com.google.common.primitives"), data.usedPackagesWithoutSpecificTypes);
  }

  @Test
  public void testAnnotationAfterImport() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("AnnotationAfterImport.java"));
    assertEquals(
        Map.of(
            "workspace.com.gazelle.java.javaparser.generators.AnnotationAfterImport",
            new PerClassData(treeSet("com.example.FlakyTest"), new TreeMap<>(), new TreeMap<>())),
        data.perClassData);
  }

  @Test
  public void testAnnotationAfterImportOnNestedClass() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("NestedClassAnnotations.java"));
    assertEquals(
        Map.of(
            "workspace.com.gazelle.java.javaparser.generators.NestedClassAnnotations.Inner",
            new PerClassData(treeSet("com.example.FlakyTest"), new TreeMap<>(), new TreeMap<>())),
        data.perClassData);
  }

  @Test
  public void testAnnotationOnField() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("AnnotationOnField.java"));

    TreeMap<String, SortedSet<String>> expectedOuterClassFieldAnnotations = new TreeMap<>();
    expectedOuterClassFieldAnnotations.put("someField", treeSet("lombok.Getter"));

    TreeMap<String, SortedSet<String>> expectedInnerClassFieldAnnotations = new TreeMap<>();
    expectedInnerClassFieldAnnotations.put("canBeSet", treeSet("lombok.Setter"));

    TreeMap<String, SortedSet<String>> expectedInnerEnumFieldAnnotations = new TreeMap<>();
    expectedInnerEnumFieldAnnotations.put("size", treeSet("lombok.Getter"));

    TreeMap<String, PerClassData> expected = new TreeMap<>();
    expected.put(
        "workspace.com.gazelle.java.javaparser.generators.AnnotationOnField",
        new PerClassData(new TreeSet<>(), new TreeMap<>(), expectedOuterClassFieldAnnotations));
    expected.put(
        "workspace.com.gazelle.java.javaparser.generators.AnnotationOnField.InnerClass",
        new PerClassData(new TreeSet<>(), new TreeMap<>(), expectedInnerClassFieldAnnotations));
    expected.put(
        "workspace.com.gazelle.java.javaparser.generators.AnnotationOnField.InnerEnum",
        new PerClassData(new TreeSet<>(), new TreeMap<>(), expectedInnerEnumFieldAnnotations));

    assertEquals(expected, data.perClassData);
  }

  @Test
  public void testAnnotationAfterImportOnMethod() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("AnnotationAfterImportOnMethod.java"));

    TreeMap<String, SortedSet<String>> expectedPerMethodAnnotations = new TreeMap<>();
    expectedPerMethodAnnotations.put("someTest", treeSet("org.junit.jupiter.api.Test"));

    assertEquals(
        Map.of(
            "workspace.com.gazelle.java.javaparser.generators.AnnotationAfterImportOnMethod",
            new PerClassData(new TreeSet<>(), expectedPerMethodAnnotations, new TreeMap<>())),
        data.perClassData);
  }

  @Test
  public void testAnnotationFromJavaStandardLibrary() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("AnnotationFromJavaStandardLibrary.java"));
    assertEquals(
        Map.of(
            "workspace.com.gazelle.java.javaparser.generators.AnnotationFromJavaStandardLibrary",
            new PerClassData(treeSet("Deprecated"), new TreeMap<>(), new TreeMap<>())),
        data.perClassData);
  }

  @Test
  public void testAnnotationWithoutImport() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("AnnotationWithoutImport.java"));
    assertEquals(
        Map.of(
            "workspace.com.gazelle.java.javaparser.generators.AnnotationWithoutImport",
            new PerClassData(treeSet("WhoKnowsWhereIAmFrom"), new TreeMap<>(), new TreeMap<>())),
        data.perClassData);
  }

  @Test
  @Disabled("Turbine has no method bodies: FQN refs used only inside methods are not captured")
  public void testFullyQualifieds() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("FullyQualifieds.java"));
    Set<String> expected =
        Set.of(
            "workspace.com.gazelle.java.javaparser.generators.DeleteBookRequest",
            "workspace.com.gazelle.java.javaparser.generators.DeleteBookResponse",
            "workspace.com.gazelle.java.javaparser.utils.Printer",
            "workspace.com.gazelle.java.javaparser.factories.Factory",
            "java.util.ArrayList");
    assertEquals(expected, data.usedTypes);
  }

  @Test
  @Disabled(
      "Turbine has no method bodies: anonymous inner class annotations in field initialisers are"
          + " not captured")
  public void testAnonymousInnerClass() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("AnonymousInnerClass.java"));

    Set<String> expectedTypes =
        Set.of(
            "java.util.HashMap", "javax.annotation.Nullable", "org.jetbrains.annotations.Nullable");
    assertEquals(expectedTypes, data.usedTypes);

    Map<String, PerClassData> expectedPerClassMetadata = new TreeMap<>();
    TreeMap<String, SortedSet<String>> expectedPerMethodAnnotations = new TreeMap<>();
    expectedPerMethodAnnotations.put(
        "containsValue", treeSet("Override", "javax.annotation.Nullable"));
    expectedPerClassMetadata.put(
        "workspace.com.gazelle.java.javaparser.generators.AnonymousInnerClass.",
        new PerClassData(treeSet(), expectedPerMethodAnnotations, new TreeMap<>()));
    assertEquals(expectedPerClassMetadata, data.perClassData);
  }

  @Test
  public void testMethodWithImportedType() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("MethodWithImportedType.java"));
    Set<String> expected =
        Set.of(
            "com.example.OuterReturnType",
            "com.example.OtherOuterReturnType",
            "com.example.Outer.InnerReturnType");
    assertEquals(expected, data.usedTypes);
  }

  @Test
  public void testInterfaceExports() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("ExportingInterface.java"));
    Set<String> expected = Set.of("example.external.NeedsExporting", "example.external.Outer");
    assertEquals(expected, data.exportedTypes);
  }

  @Test
  public void testClassExports() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("ExportingClass.java"));
    Set<String> expected =
        Set.of(
            "example.external.PackageReturn",
            "example.external.ParameterizedArgReturn",
            "example.external.ParameterizedOuterReturn",
            "example.external.ProtectedReturn",
            "example.external.PublicReturn");
    assertEquals(expected, data.exportedTypes);
  }

  @Test
  public void testParameterExports() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("ParameterExportingClass.java"));
    Set<String> expected =
        Set.of(
            "example.external.ConstructorParam",
            "example.external.PackageParam",
            "example.external.ParameterizedArg",
            "example.external.ParameterizedParam",
            "example.external.ProtectedParam",
            "example.external.PublicParam",
            "example.external.VarargParam",
            "example.external.WildcardBound");
    assertEquals(expected, data.exportedTypes);
  }

  @Test
  public void testFieldExports() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("FieldExportingClass.java"));
    Set<String> expected =
        Set.of(
            "example.external.PackageField",
            "example.external.ParameterizedFieldArg",
            "example.external.ParameterizedFieldOuter",
            "example.external.ProtectedField",
            "example.external.PublicField");
    assertEquals(expected, data.exportedTypes);
  }

  @Test
  public void testSamePackageBareClassReference() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("SamePackageReference.java"));
    assertEquals(
        Set.of(
            "workspace.com.gazelle.java.javaparser.generators.AbstractIdentifier",
            "workspace.com.gazelle.java.javaparser.generators.SomeHelper",
            "workspace.com.gazelle.java.javaparser.generators.SomeInput",
            "workspace.com.gazelle.java.javaparser.generators.SomeInterface"),
        data.usedTypes);
  }

  @Test
  public void testSamePackageFiltersInnerClasses() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("SamePackageWithInnerClass.java"));
    assertEquals(
        Set.of("workspace.com.gazelle.java.javaparser.generators.ExternalHelper"), data.usedTypes);
  }

  @Test
  public void testSamePackageFiltersTypeParameters() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("SamePackageWithGenerics.java"));
    assertEquals(
        Set.of("workspace.com.gazelle.java.javaparser.generators.SomeBound"), data.usedTypes);
  }

  @Test
  public void testSamePackageAllTypePositions() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("SamePackageAllPositions.java"));
    assertEquals(
        Set.of(
            "workspace.com.gazelle.java.javaparser.generators.SomeClassAnnotation",
            "workspace.com.gazelle.java.javaparser.generators.SomeFieldType",
            "workspace.com.gazelle.java.javaparser.generators.SomeMethodAnnotation",
            "workspace.com.gazelle.java.javaparser.generators.SomeMethodBound",
            "workspace.com.gazelle.java.javaparser.generators.SomeCheckedException"),
        data.usedTypes);
  }

  @Test
  public void testFqnAnnotationOnFieldAndClass() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("FqnAnnotationOnFieldAndClass.java"));
    assertEquals(
        Set.of("com.example.ClassAnnotation", "com.example.FieldAnnotation"), data.usedTypes);
  }

  @Test
  @Disabled("Turbine has no method bodies: Foo.class literals inside methods are not captured")
  public void testClassLiteral() throws IOException {
    ParsedPackageData data = parser.parseClasses(directory, List.of("ClassLiteral.java"));
    assertEquals(
        Set.of(
            "com.example.Registry", "workspace.com.gazelle.java.javaparser.generators.MyHandler"),
        data.usedTypes);
  }

  @Test
  @Disabled(
      "Turbine has no method bodies: bare method receivers (SomeClass.method()) inside methods are"
          + " not captured")
  public void testBareClassMethodReceiver() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("BareClassMethodReceiver.java"));
    assertEquals(
        Set.of("workspace.com.gazelle.java.javaparser.generators.SamePackageHelper"),
        data.usedTypes);
  }

  @Test
  public void testStaticImportNestedClass() throws IOException {
    ParsedPackageData data =
        parser.parseClasses(directory, List.of("StaticImportNestedClass.java"));
    assertEquals(Set.of("com.example.Outer", "com.example.Outer.Inner"), data.usedTypes);
  }

  @Test
  public void parseClassesByPath(@TempDir Path tempDir) throws IOException {
    Path src = tempDir.resolve("Greeter.java");
    Files.writeString(src, "package demo; public class Greeter {}");
    ParsedPackageData data = parser.parseClasses(tempDir, List.of("Greeter.java"));
    assertEquals(Set.of("demo"), data.packages);
  }

  @SafeVarargs
  private <T> TreeSet<T> treeSet(T... values) {
    return new TreeSet<>(Arrays.asList(values));
  }
}

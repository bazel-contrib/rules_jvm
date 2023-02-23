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
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeSet;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import javax.tools.JavaFileObject;
import javax.tools.SimpleJavaFileObject;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ClasspathParserTest {
  private static final Logger logger = LoggerFactory.getLogger(ClasspathParserTest.class);

  private static Path workspace;
  private static Path directory;

  private ClasspathParser parser;
  private static Map<String, ? extends JavaFileObject> testFiles;

  @BeforeAll
  @SuppressFBWarnings("RCN_REDUNDANT_NULLCHECK_WOULD_HAVE_BEEN_A_NPE") // See
  // https://github.com/spotbugs/spotbugs/issues/1694
  public static void setup() throws IOException, URISyntaxException {
    URI uri = ClasspathParserTest.class.getClassLoader().getResource("workspace").toURI();
    Map<String, String> env = new HashMap<>();
    env.put("create", "true");
    FileSystems.newFileSystem(uri, env);
    workspace = Paths.get(uri);
    directory = workspace.resolve("com/gazelle/java/javaparser/generators");
    try (Stream<Path> stream = Files.list(directory)) {
      testFiles =
          stream
              .filter(file -> !Files.isDirectory(file))
              .map(name -> new JavaSource(name, name.toString()))
              .collect(Collectors.toMap(SimpleJavaFileObject::getName, source -> source));
    }
    logger.info("Got Test Files {}", testFiles);
  }

  @BeforeEach
  public void setupPerTest() {
    parser = new ClasspathParser();
  }

  @Test
  public void pathTest() {
    logger.info("WORKSPACE={}", workspace);
    assertEquals("/workspace", workspace.toString());
    logger.info("directory={}", directory);
    assertEquals("/workspace/com/gazelle/java/javaparser/generators", directory.toString());
  }

  @Test
  public void simpleTest() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(testFiles.get("/workspace/com/gazelle/java/javaparser/generators/Main.java"));
    assertEquals(1, files.size());
    parser.parseClasses(files);

    Assertions.assertTrue(parser.getUsedTypes().isEmpty());
    assertEquals(Set.of("workspace.com.gazelle.java.javaparser.generators"), parser.getPackages());
    assertEquals(Set.of("Main"), parser.getMainClasses());
  }

  @Test
  public void verifyPackages() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(testFiles.get("/workspace/com/gazelle/java/javaparser/generators/Main.java"));
    parser.parseClasses(files);
    assertEquals(Set.of("workspace.com.gazelle.java.javaparser.generators"), parser.getPackages());
  }

  @Test
  public void verifyMainClasses() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(testFiles.get("/workspace/com/gazelle/java/javaparser/generators/Main.java"));
    parser.parseClasses(files);

    assertEquals(Set.of("Main"), parser.getMainClasses());
  }

  @Test
  public void verifyNoMainClasses() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(
            testFiles.get(
                "/workspace/com/gazelle/java/javaparser/generators/ClasspathParser.java"));
    parser.parseClasses(files);

    assertEquals(0, parser.getMainClasses().size());
  }

  @Test
  public void verifyPackagesUnique() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(
            testFiles.get("/workspace/com/gazelle/java/javaparser/generators/Main.java"),
            testFiles.get("/workspace/com/gazelle/java/javaparser/generators/PackageParser.java"),
            testFiles.get(
                "/workspace/com/gazelle/java/javaparser/generators/ClasspathParser.java"));
    parser.parseClasses(files);

    assertEquals(Set.of("workspace.com.gazelle.java.javaparser.generators"), parser.getPackages());
  }

  @Test
  public void verifyImportsOnParse() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(testFiles.get("/workspace/com/gazelle/java/javaparser/generators/Hello.java"));
    parser.parseClasses(files);

    assertEquals(
        Set.of(
            "com.google.common.primitives.Ints",
            "workspace.com.gazelle.java.javaparser.generators.DeleteBookRequest",
            "workspace.com.gazelle.java.javaparser.generators.HelloProto"),
        parser.getUsedTypes());
  }

  @Test
  public void testWildcardImportOverlap() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(testFiles.get("/workspace/com/gazelle/java/javaparser/generators/Wildcards.java"));
    parser.parseClasses(files);
    assertEquals(Set.of("org.junit.jupiter.api", "org.junit.jupiter.api.Assertions"), parser.getUsedTypes());
  }

  @Test
  public void testFullyQualifiedClassUseNotViaImport() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(
            testFiles.get("/workspace/com/gazelle/java/javaparser/generators/PackageParser.java"));
    parser.parseClasses(files);
    assertEquals(
        Set.of(
            "com.gazelle.java.ArrayParse",
            "com.gazelle.java.ClasspathParser",
            "com.gazelle.java.OtherClasspathParse"),
        parser.getUsedTypes());
  }

  @Test
  public void testStaticImport() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(
            testFiles.get("/workspace/com/gazelle/java/javaparser/generators/StaticImports.java"));
    parser.parseClasses(files);

    assertEquals(Set.of("com.gazelle.java.javaparser.ClasspathParser"), parser.getUsedTypes());
  }

  @Test
  public void testWildcardImport() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(
            testFiles.get("/workspace/com/gazelle/java/javaparser/generators/WildcardImport.java"));
    parser.parseClasses(files);

    assertEquals(Set.of("com.google.common.primitives"), parser.getUsedTypes());
  }

  @Test
  public void testAnnotationAfterImport() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(
            testFiles.get(
                "/workspace/com/gazelle/java/javaparser/generators/AnnotationAfterImport.java"));
    parser.parseClasses(files);

    assertEquals(
        Map.of(
            "workspace.com.gazelle.java.javaparser.generators.AnnotationAfterImport",
            treeSet("com.example.FlakyTest")),
        parser.getAnnotatedClasses());
  }

  @Test
  public void testAnnotationFromJavaStandardLibrary() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(
            testFiles.get(
                "/workspace/com/gazelle/java/javaparser/generators/AnnotationFromJavaStandardLibrary.java"));
    parser.parseClasses(files);

    // Ideally this would resolve to java.lang.Deprecated, but nothing currently does that
    // resolution.
    assertEquals(
        Map.of(
            "workspace.com.gazelle.java.javaparser.generators.AnnotationFromJavaStandardLibrary",
            treeSet("Deprecated")),
        parser.getAnnotatedClasses());
  }

  @Test
  public void testAnnotationWithoutImport() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(
            testFiles.get(
                "/workspace/com/gazelle/java/javaparser/generators/AnnotationWithoutImport.java"));
    parser.parseClasses(files);

    // Ideally this would resolve to a fully-qualified class-name, but we don't currently keep
    // enough state to do that resolution, so we report what we can.
    assertEquals(
        Map.of(
            "workspace.com.gazelle.java.javaparser.generators.AnnotationWithoutImport",
            treeSet("WhoKnowsWhereIAmFrom")),
        parser.getAnnotatedClasses());
  }

  @Test
  public void testFullyQualifieds() throws IOException {
    List<? extends JavaFileObject> files =
        List.of(
            testFiles.get(
                "/workspace/com/gazelle/java/javaparser/generators/FullyQualifieds.java"));
    parser.parseClasses(files);

    Set<String> expected =
        Set.of(
            "workspace.com.gazelle.java.javaparser.generators.DeleteBookRequest",
            "workspace.com.gazelle.java.javaparser.generators.DeleteBookResponse",
            "workspace.com.gazelle.java.javaparser.utils.Printer",
            "workspace.com.gazelle.java.javaparser.factories.Factory");
    assertEquals(expected, parser.getUsedTypes());
  }

  private <T> TreeSet<T> treeSet(T... values) {
    TreeSet<T> set = new TreeSet<>();
    for (T value : values) {
      set.add(value);
    }
    return set;
  }

  static class JavaSource extends SimpleJavaFileObject {
    String fileSource;

    public JavaSource(Path directory, String fileName) {
      super(Path.of(fileName).toUri(), JavaFileObject.Kind.SOURCE);
      readFileFromSource(directory);
    }

    private void readFileFromSource(Path fileName) {
      try {
        fileSource = Files.readString(fileName);
      } catch (IOException ex) {
        logger.error("Unable to read build file: {} : {}", fileName, ex.getMessage());
      }
    }

    @Override
    public CharSequence getCharContent(boolean ignoreEncodingErrors) {
      return fileSource;
    }
  }
}

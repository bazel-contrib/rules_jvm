package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;

import com.github.javaparser.JavaParser;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.nio.file.FileSystems;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ClasspathParserTest {
  private static final Logger logger = LoggerFactory.getLogger(ClasspathParserTest.class);

  private static Path workspace;
  private static JavaParser javaParser;
  private static Path directory;

  private ClasspathParser parser;

  @BeforeAll
  public static void setup() throws IOException, URISyntaxException {
    URI uri = ClasspathParserTest.class.getClassLoader().getResource("workspace").toURI();
    Map<String, String> env = new HashMap<>();
    env.put("create", "true");
    FileSystems.newFileSystem(uri, env);
    workspace = Paths.get(uri);
    directory = workspace.resolve("com/gazelle/java/javaparser/generators");
    PackageParser parser = new PackageParser(workspace);
    parser.setup(".", ".", null);
    javaParser = parser.getJavaParser();
  }

  @BeforeEach
  public void setupPerTest() {
    parser = new ClasspathParser(ClasspathParserTest.javaParser);
  }

  @Test
  public void pathTest() {
    logger.info("WORKSPACE={}", workspace);
    assertEquals("/workspace", workspace.toString());
    logger.info("directory={}", directory);
    assertEquals("/workspace/com/gazelle/java/javaparser/generators", directory.toString());
  }

  @Test
  public void simpleTest() {
    List<String> files = List.of("Main.java");

    parser.parseClasses(workspace, directory, files);
    assertFalse(parser.getUsedTypes().isEmpty());
    assertFalse(parser.getPackages().isEmpty());
    assertFalse(parser.getMainClasses().isEmpty());
  }

  @Test
  public void verifyPackages() {
    List<String> files = List.of("Main.java");

    parser.parseClasses(workspace, directory, files);

    List<String> packages = new ArrayList<>(parser.getPackages());

    assertEquals(1, packages.size());
    assertEquals("workspace.com.gazelle.java.javaparser.generators", packages.get(0));
  }

  @Test
  public void verifyMainClasses() {
    List<String> files = List.of("Main.java");

    parser.parseClasses(workspace, directory, files);
    List<String> mainClasses = new ArrayList<>(parser.getMainClasses());

    assertEquals(1, mainClasses.size());
    assertEquals("Main", mainClasses.get(0));
  }

  @Test
  public void verifyNoMainClasses() {
    List<String> files = List.of("ClasspathParser.java");

    parser.parseClasses(workspace, directory, files);
    List<String> mainClasses = new ArrayList<>(parser.getMainClasses());

    assertEquals(0, mainClasses.size());
  }

  @Test
  public void verifyPackagesUnique() {
    List<String> files = List.of("Main.java", "ClasspathParser.java", "PackageParser.java");

    parser.parseClasses(workspace, directory, files);
    List<String> packages = new ArrayList<>(parser.getPackages());

    assertEquals(1, packages.size());
    assertEquals("workspace.com.gazelle.java.javaparser.generators", packages.get(0));
  }

  @Test
  public void verifyImportsOnParse() {
    List<String> files = List.of("Hello.java");

    parser.parseClasses(workspace, directory, files);
    List<String> types = new ArrayList<>(parser.getUsedTypes());

    assertArrayEquals(
        types.toArray(),
        List.of(
                "com.google.common.primitives.Ints",
                "workspace.com.gazelle.java.javaparser.generators.DeleteBookRequest",
                "workspace.com.gazelle.java.javaparser.generators.HelloProto")
            .toArray());
  }

  @Test
  public void verifyImportsForSelf() {
    List<String> files = List.of("App.java");

    parser.parseClasses(workspace, directory, files);
    List<String> types = new ArrayList<>(parser.getUsedTypes());

    assertArrayEquals(
        types.toArray(),
        List.of(
                "com.google.common.primitives.Ints",
                "java.lang.Exception",
                "java.lang.String",
                "workspace.com.gazelle.java.javaparser.generators.App")
            .toArray());
  }

  @Test
  public void testForMainInSelf() {
    List<String> files = List.of("App.java");

    parser.parseClasses(workspace, directory, files);
    List<String> mainClasses = new ArrayList<>(parser.getMainClasses());

    assertEquals(1, mainClasses.size());
    assertEquals("App", mainClasses.get(0));
  }

  @Test
  public void testWildcardImport() {
    List<String> files = List.of("Wildcards.java");
    parser.parseClasses(workspace, directory, files);
    List<String> types = new ArrayList<>(parser.getUsedTypes());

    assertArrayEquals(types.toArray(), List.of("org.junit.jupiter.api.Assertions").toArray());
  }
}

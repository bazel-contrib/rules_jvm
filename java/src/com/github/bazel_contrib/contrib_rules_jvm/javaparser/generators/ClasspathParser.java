package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import com.github.javaparser.JavaParser;
import com.github.javaparser.ParseProblemException;
import com.github.javaparser.ParseResult;
import com.github.javaparser.ast.CompilationUnit;
import com.github.javaparser.ast.ImportDeclaration;
import com.github.javaparser.ast.PackageDeclaration;
import com.github.javaparser.ast.body.ClassOrInterfaceDeclaration;
import com.github.javaparser.ast.body.MethodDeclaration;
import com.github.javaparser.ast.type.ArrayType;
import com.github.javaparser.ast.type.ClassOrInterfaceType;
import com.github.javaparser.resolution.UnsolvedSymbolException;
import com.github.javaparser.resolution.declarations.ResolvedReferenceTypeDeclaration;
import com.github.javaparser.resolution.types.ResolvedReferenceType;
import java.io.IOException;
import java.nio.file.FileSystems;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.PathMatcher;
import java.util.List;
import java.util.Set;
import java.util.TreeSet;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ClasspathParser {
  private static final Logger logger = LoggerFactory.getLogger(ClasspathParser.class);

  private final Set<String> usedTypes = new TreeSet<>();
  private final Set<String> packages = new TreeSet<>();
  private final Set<String> mainClasses = new TreeSet<>();
  private final JavaParser javaParser;

  public ClasspathParser(JavaParser javaParser) {
    this.javaParser = javaParser;
  }

  public Set<String> getUsedTypes() {
    return usedTypes;
  }

  public Set<String> getPackages() {
    return packages;
  }

  public Set<String> getMainClasses() {
    return mainClasses;
  }

  public void parseClasses(String srcs, Path workspace) throws IOException {
    String pattern = workspace.toString() + "/" + srcs;
    logger.debug("Pattern: {}", pattern);
    PathMatcher pathMatcher = FileSystems.getDefault().getPathMatcher("glob:" + pattern);
    try (Stream<Path> paths =
        Files.find(workspace, Integer.MAX_VALUE, (path, f) -> pathMatcher.matches(path))) {
      paths
          .peek(path -> logger.debug("processing {}", path))
          .forEach(this::parseFileGatherDependencies);
    }
  }

  public void parseClasses(Path workspace, Path directory, List<String> files) {
    files.stream()
        .map(filename -> workspace.resolve(directory).resolve(filename))
        .forEach(this::parseFileGatherDependencies);
  }

  private void parsePackages(CompilationUnit cu) {
    packages.addAll(
        cu.findAll(PackageDeclaration.class).stream()
            .map(PackageDeclaration::getNameAsString)
            .collect(Collectors.toList()));
  }

  private void findMainMethods(CompilationUnit cu) {
    List<String> mains =
        cu.findAll(MethodDeclaration.class).stream()
            .filter(MethodDeclaration::isStatic)
            .filter(MethodDeclaration::isPublic)
            .filter(m -> m.getType().isVoidType())
            .filter(m -> m.getNameAsString().equals("main"))
            .map(m -> ((ClassOrInterfaceDeclaration) m.getParentNode().get()).getNameAsString())
            .collect(Collectors.toList());
    if (!mains.isEmpty()) {
      mainClasses.addAll(mains);
    }
  }

  private void parseImports(CompilationUnit cu) {
    // IMPORTS : Checking the imports
    cu.findAll(ImportDeclaration.class)
        .forEach(
            id -> {
              String name = id.getNameAsString();
              if (id.isAsterisk()) {
                logger.debug("Handling wildcard import: {} to package {}", id, name);
                usedTypes.add(name);
              } else if (id.isStatic()) {
                String staticPackage = name.substring(0, name.lastIndexOf('.'));
                usedTypes.add(staticPackage);
              } else {
                usedTypes.add(name);
              }
            });
  }

  private void parseClasses(CompilationUnit cu) {
    // CLASSES : Checking the fully qualified class or interface names
    cu.findAll(ClassOrInterfaceType.class)
        .forEach(
            coit -> {
              ResolvedReferenceTypeDeclaration type;
              String typeName = null;
              if (!Character.isUpperCase(coit.getName().asString().charAt(0))) {
                logger.trace(
                    "Working on {} and thinking this is a package, so skipping",
                    coit.getName().asString());
              } else if (coit.isArrayType()) {
                ArrayType arrayTyoe = coit.asArrayType();
                type =
                    arrayTyoe
                        .resolve()
                        .getComponentType()
                        .asReferenceType()
                        .getTypeDeclaration()
                        .get();
                typeName = type.getQualifiedName();
              } else {
                try {
                  ResolvedReferenceType ref = coit.resolve();
                  type = ref.getTypeDeclaration().get();
                  typeName = type.getQualifiedName();
                } catch (UnsolvedSymbolException exception) {
                  logger.trace(
                      "Working on class {} And unable to find: {} -" + " Continuing",
                      cu.getPrimaryTypeName().get(),
                      exception.getName());
                } catch (UnsupportedOperationException exception) {
                  // The ResolvedReferenceType is the generics for some classes,
                  // which the system
                  // can not resolve
                  // begin generic. We're not alerting on this, but skipping and
                  // assuming we don't
                  // need to resolve
                  // the generic type reference by applying a dependency.
                  if (exception.getMessage().contains("CorrespondingDeclaration")) {
                    logger.debug(
                        "Working on classes from {} - {}\n"
                            + " unable to find generic  - Continuing",
                        cu.getPrimaryTypeName().get(),
                        coit.getName());
                  } else if (!exception.getMessage().contains("ResolvedReferenceType")) {
                    logger.error(
                        "Working on classes from {} - {}\n" + " Caught exception :",
                        cu.getPrimaryTypeName().get(),
                        coit.getName(),
                        exception);
                    throw exception;
                  }
                }
              }
              if (typeName != null) {
                usedTypes.add(typeName);
              }
            });
  }

  private void parseFileGatherDependencies(Path classPath) {
    CompilationUnit cu;
    try {
      ParseResult<CompilationUnit> result = this.javaParser.parse(classPath);
      if (result.isSuccessful()) {
        cu = result.getResult().get();
      } else {
        throw new ParseProblemException(result.getProblems());
      }
    } catch (IOException exception) {
      logger.warn("Unable to parse {}. Skipping File", classPath);
      return;
    }
    parseImports(cu);
    parseClasses(cu);
    parsePackages(cu);
    findMainMethods(cu);
  }
}

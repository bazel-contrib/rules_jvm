package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import com.github.bazel_contrib.contrib_rules_jvm.javaparser.file.BuildFile;
import com.github.javaparser.JavaParser;
import com.github.javaparser.ParseProblemException;
import com.github.javaparser.ParseResult;
import com.github.javaparser.ast.CompilationUnit;
import com.github.javaparser.ast.ImportDeclaration;
import com.github.javaparser.ast.body.Parameter;
import com.github.javaparser.ast.type.ArrayType;
import com.github.javaparser.ast.type.ClassOrInterfaceType;
import com.github.javaparser.resolution.UnsolvedSymbolException;
import com.github.javaparser.resolution.declarations.ResolvedReferenceTypeDeclaration;
import com.github.javaparser.resolution.types.ResolvedReferenceType;
import java.io.IOException;
import java.nio.file.Path;
import java.util.List;
import java.util.Objects;
import java.util.Set;
import java.util.TreeSet;
import java.util.stream.Collectors;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class PackageDependencies {
  private static final Logger logger = LoggerFactory.getLogger(PackageDependencies.class);

  private final BuildFile buildFile;
  private final KnownTypeSolvers typeSolver;
  private JavaParser javaParser;

  /** Set of java types used by all the java files in this package (Bazel build target). */
  Set<JavaIdentifier> usedTypes = new TreeSet<>();

  public PackageDependencies(BuildFile buildFile, KnownTypeSolvers typeSolver) {
    this.buildFile = buildFile;
    this.typeSolver = typeSolver;
  }

  public String toString() {
    return usedTypes.toString();
  }

  public Path getPackagePath() {
    return buildFile.getBuildPath();
  }

  public String getTargetName() {
    return buildFile.getTargetName();
  }

  public BuildFile getBuildFile() {
    return buildFile;
  }

  public void setJavaParser(JavaParser javaParser) {
    Objects.requireNonNull(javaParser);
    this.javaParser = javaParser;
  }

  public void resolveTypesForClass(Path classPath) {
    CompilationUnit cu;

    try {
      ParseResult<CompilationUnit> result = javaParser.parse(classPath);
      if (result.isSuccessful()) {
        cu = result.getResult().get();
      } else {
        throw new ParseProblemException(result.getProblems());
      }
    } catch (IOException exception) {
      logger.error("Unable to parse {}. Skipping File", classPath);
      return;
    }
    try {
      resolveImports(cu);
    } catch (UnsolvedSymbolException ex) {
      logger.warn(
          "Working on imports from {}\n   And unable to find: {}\n"
              + "You need to add this dependency to your overall dependency list.",
          classPath,
          ex.getName());
    } catch (UnsupportedOperationException | ParseProblemException exception) {
      logger.error("Working on imports from {}\n Caught exception: {}", classPath, exception);
    }
    try {
      resolveClasses(cu);
    } catch (UnsolvedSymbolException ex) {
      logger.warn(
          "Working on classes from {}\n"
              + "   And unable to find: {}\n"
              + "    You need to add this dependency to your overall dependency list.",
          classPath,
          ex.getName());
    } catch (UnsupportedOperationException | ParseProblemException exception) {
      logger.error("Working on classes from {}\n Caught exception: {}", classPath, exception);
    }
    resolveParameters(cu);
  }

  public String updateBuildFile(boolean dryRun) throws IOException {
    String deps = generateDeps();
    String update = buildFile.replaceTarget("_deps", deps);
    if (!dryRun) {
      buildFile.writeUpdatedFile(update);
    }
    return update;
  }

  public String generateDeps() {
    List<String> libraries = generateTargetsList();

    StringBuilder targetRule = new StringBuilder();
    targetRule.append(buildFile.getTargetName());
    targetRule.append("_deps = [\n");
    for (String library : libraries) {
      targetRule.append("    ");
      targetRule.append(library);
      targetRule.append("\n");
    }
    targetRule.append("]\n\n");

    return targetRule.toString();
  }

  private void resolveImports(CompilationUnit cu) {
    cu.findAll(ImportDeclaration.class)
        .forEach(
            id -> {
              if (id.isAsterisk()) {
                throw new UnsupportedOperationException(
                    "Haven't figured out what to do with wildcard imports");
              } else if (id.isStatic()) {
                String name = id.getNameAsString();
                String staticPackage = name.substring(0, name.lastIndexOf('.'));
                usedTypes.add(typeSolver.resolveSourceForType(staticPackage));
              } else {
                usedTypes.add(typeSolver.resolveSourceForType(id.getNameAsString()));
              }
            });
  }

  private void resolveClasses(CompilationUnit cu) {
    cu.findAll(ClassOrInterfaceType.class)
        .forEach(
            coit -> {
              ResolvedReferenceTypeDeclaration type;
              ResolvedReferenceType ref = null;
              String typeName = null;
              if (!Character.isUpperCase(coit.getName().asString().charAt(0))) {
                logger.debug(
                    "Working on {} and thinking this is a package, so" + " skippling",
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
                typeName = type.getPackageName() + "." + type.getClassName();
              } else {
                try {
                  ref = coit.resolve();
                } catch (UnsolvedSymbolException exception) {
                  logger.error(
                      "Working on class {} And unable to find: {}",
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
                  if (!exception.getMessage().contains("ResolvedReferenceType")) {
                    logger.error(
                        "Working on classes from {}\n Caught exception :{}",
                        cu.getPrimaryTypeName().get(),
                        exception);
                    throw exception;
                  }
                } catch (IllegalStateException exception) {
                  logger.error(
                      "Working on class{}\n Caught Exception: {} ",
                      cu.getPrimaryTypeName().get(),
                      exception.getMessage());
                  throw exception;
                }
              }
              if (ref != null) {
                type = ref.getTypeDeclaration().get();
                typeName = type.getPackageName() + "." + type.getClassName();
              }
              if (typeName != null) {
                JavaIdentifier source = typeSolver.resolveSourceForType(typeName);
                usedTypes.add(source);
              }
            });
  }

  private void resolveParameters(CompilationUnit cu) {
    cu.findAll(Parameter.class)
        .forEach(
            p -> {
              //            System.out.println(p);
            });
  }

  /**
   * Generate a list of unique, sorted bazel dependency strings from the libraries used by this
   * package.
   *
   * @return the list of dependencies.
   */
  private List<String> generateTargetsList() {
    return usedTypes.stream()
        .map(JavaIdentifier::getSourceLibrary)
        .filter(Objects::nonNull)
        .filter(dep -> !dep.equals(buildFile.getBazelTarget()))
        .distinct()
        .sorted()
        .collect(Collectors.toList());
  }
}

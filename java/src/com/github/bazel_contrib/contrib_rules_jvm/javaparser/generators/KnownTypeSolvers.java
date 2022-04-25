package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import com.github.javaparser.resolution.UnsolvedSymbolException;
import com.github.javaparser.resolution.declarations.ResolvedReferenceTypeDeclaration;
import com.github.javaparser.symbolsolver.JavaSymbolSolver;
import com.github.javaparser.symbolsolver.model.resolution.SymbolReference;
import com.github.javaparser.symbolsolver.model.resolution.TypeSolver;
import com.github.javaparser.symbolsolver.resolution.typesolvers.CombinedTypeSolver;
import com.github.javaparser.symbolsolver.resolution.typesolvers.JavaParserTypeSolver;
import com.github.javaparser.symbolsolver.resolution.typesolvers.ReflectionTypeSolver;
import com.google.common.annotations.VisibleForTesting;
import java.io.File;
import java.io.IOException;
import java.nio.file.FileSystems;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.PathMatcher;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class KnownTypeSolvers {
  private static final Logger logger = LoggerFactory.getLogger(KnownTypeSolvers.class);
  /**
   * This is a map of a Java Type solver, the class that will look up (resolve or solve) a given
   * java type, to the bazel dependency name where Bazel will find the required java type. The
   * dependency name will be of the form: - artifact("com.example.library:library") - For a
   * dependency external to the repo and found in the maven cache - //src/java/ - For a dependency
   * in the repo the root of thw source tree for searching - null - For all dependencies in the
   * default java library
   *
   * <p>The complete map will contain a ReflectionTypeSolver (for the java library), one type solver
   * for each jar file in the maven cache, and one type solver for each source project directory in
   * your repo, each with the corresponding bazel dependency name.
   */
  private final Map<TypeSolver, String> knownSolvers = new HashMap<>();

  /**
   * A collection of the same solvers in the knownSolvers. Used by the JavaParser internally for
   * type resolution during parsing.
   */
  private final CombinedTypeSolver typeSolver;

  /** List of source/test packages used for */
  private final List<PackageDependencies> packages;

  public KnownTypeSolvers(List<PackageDependencies> packages) {
    typeSolver = new CombinedTypeSolver(new ReflectionTypeSolver());
    knownSolvers.put(new ReflectionTypeSolver(), null);
    this.packages = packages;
  }

  public JavaSymbolSolver getTypeSolver() {
    return new JavaSymbolSolver(typeSolver);
  }

  /**
   * Resolve the location of a Java type to a specific identifier. Start with a specific java type
   * (e.g. com.gazelle.java.javaparser.generators.Main) and using the collected set of type solvers,
   * figure out which jar file or source tree has the type.
   *
   * @param typeName Type name from the Java Parser (e.g.
   *     com.gazelle.java.javaparser.generators.Main)
   * @return JavaIdentifier of the package and source location
   */
  public JavaIdentifier resolveSourceForType(String typeName) {
    for (Map.Entry<TypeSolver, String> pair : knownSolvers.entrySet()) {
      TypeSolver solver = pair.getKey();
      String sourceLibrary = pair.getValue();
      SymbolReference<ResolvedReferenceTypeDeclaration> ref = solver.tryToSolveType(typeName);
      if (ref.isSolved()) {
        ResolvedReferenceTypeDeclaration res = ref.getCorrespondingDeclaration();
        if (sourceLibrary != null && sourceLibrary.startsWith("//")) {
          sourceLibrary =
              getSourceBazelTarget(sourceLibrary, Arrays.asList(res.getPackageName().split("\\.")));
        }
        // Return the the entry code of where this was found to mark as a value
        // Note: This assumes every type is unique in the whole of the library set
        // and if not returns one library source at random.
        return new JavaIdentifier(res.getPackageName(), res.getClassName(), sourceLibrary);
      }
      // Not a solved type, we just keep going with the next type solver in the list
    }
    // nothing found, throw an exception
    throw new UnsolvedSymbolException(typeName);
  }

  /**
   * Get the dependency resolution solvers for the code in your repository. These are the internal
   * packages kept as source in your repository (presumably your code).
   *
   * <p>The JavaParserTypeSolver is a root of the tree for a set of packages the solver will search
   * to find a given java class.
   *
   * <p>For example, if the solver is given "project/src/main" and asked to solve
   * "com.gazelle.java.javaparser.generator.Main, it converts the latter to a path, and attempts to
   * find the "Main.java" file.
   *
   * <p>The path generated for the known solvers would be the start of the bazel target path. From
   * our example "//src/main" The rest of the target path is generated when the solver is used to
   * resolve a class, based on the location of the build file.
   *
   * <p>For the simple project (like the build file generator) with only one source root there will
   * be only one internal solver. For more complicated project divided into different projects,
   * there may be several: one for each source tree project in the repository.
   *
   * @param workspace Absolute path of the workspace root of your project
   * @param srcs A path regex of project source roots.
   * @throws IOException if reading the directories causes a problem
   */
  public void getInternalSolvers(Path workspace, String srcs) throws IOException {
    String pattern = workspace.toString() + File.separatorChar + srcs;
    PathMatcher pathMatcher = FileSystems.getDefault().getPathMatcher("glob:" + pattern);
    logger.debug("Pattern: {}", pattern);
    try (Stream<Path> paths =
        Files.find(workspace, Integer.MAX_VALUE, (path, f) -> pathMatcher.matches(path))) {
      paths.forEach(
          path -> {
            JavaParserTypeSolver srcResolver = new JavaParserTypeSolver(path);
            typeSolver.add(srcResolver);
            knownSolvers.put(srcResolver, "//" + workspace.relativize(path).toString());
            logger.debug("Path added : {}", workspace.relativize(path));
          });
    }
  }

  @VisibleForTesting
  String getSourceBazelTarget(String sourceLibrary, List<String> packagePathElements) {
    // sourceLibrary -> //foo/src/main/java
    // packagePathElements -> "com", "gazelle", "java", "javaparser", "generators"
    List<String> sourcePath = Arrays.asList(sourceLibrary.split(File.separator));
    List<String> fullPath = new ArrayList<>();
    fullPath.addAll(sourcePath.subList(2, sourcePath.size()));
    fullPath.addAll(packagePathElements);
    Path packagePath = Path.of("", fullPath.toArray(new String[0]));

    List<PackageDependencies> potentials = new ArrayList<>();

    PackageDependencies match = null;
    for (PackageDependencies pkg : packages) {
      Path target = pkg.getPackagePath().getParent();
      if (target.endsWith(packagePath)) {
        match = pkg;
        break;
      } else {
        for (Path parentPath = packagePath;
            parentPath.getNameCount() > 1;
            parentPath = parentPath.getParent()) {
          if (target.endsWith(parentPath)) {
            potentials.add(pkg);
          }
        }
      }
    }
    if (match != null) {
      return match.getBuildFile().getBazelTarget();
    } else if (potentials.size() == 1) {
      return potentials.get(0).getBuildFile().getBazelTarget();
    } else if (potentials.size() > 1) {
      logger.warn("Found more than one potential target for Source files.");
      return potentials.get(0).getBuildFile().getBazelTarget();
    } else {
      return null;
    }
  }
}

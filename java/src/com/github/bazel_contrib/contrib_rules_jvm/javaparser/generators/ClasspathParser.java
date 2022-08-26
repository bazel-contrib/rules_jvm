package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static javax.lang.model.element.Modifier.PUBLIC;
import static javax.lang.model.element.Modifier.STATIC;

import com.google.common.collect.Lists;
import com.sun.source.tree.ArrayTypeTree;
import com.sun.source.tree.ClassTree;
import com.sun.source.tree.CompilationUnitTree;
import com.sun.source.tree.ImportTree;
import com.sun.source.tree.PackageTree;
import com.sun.source.tree.ParameterizedTypeTree;
import com.sun.source.tree.PrimitiveTypeTree;
import com.sun.source.tree.Tree;
import com.sun.source.tree.VariableTree;
import com.sun.source.util.JavacTask;
import com.sun.source.util.TreeScanner;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;
import java.util.Set;
import java.util.TreeSet;
import java.util.stream.Collectors;
import javax.annotation.Nullable;
import javax.lang.model.type.TypeKind;
import javax.tools.JavaCompiler;
import javax.tools.JavaFileObject;
import javax.tools.StandardJavaFileManager;
import javax.tools.ToolProvider;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ClasspathParser {
  private static final Logger logger = LoggerFactory.getLogger(ClasspathParser.class);

  private final Set<String> usedTypes = new TreeSet<>();
  private final Set<String> packages = new TreeSet<>();
  private final Set<String> mainClasses = new TreeSet<>();

  // get the system java compiler instance
  private static final JavaCompiler compiler = ToolProvider.getSystemJavaCompiler();
  private static final List<String> OPTIONS = List.of("--release=" + Runtime.version().feature());

  public ClasspathParser() {
    // Doesn't need to do anything currently
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

  public void parseClasses(Path directory, List<String> files) throws IOException {
    StandardJavaFileManager fileManager = compiler.getStandardFileManager(null, null, null);
    List<? extends JavaFileObject> objectFiles =
        files.stream()
            .map(directory::resolve)
            .map(fileName -> fileManager.getJavaFileObjects(fileName.toString()))
            .map(Lists::newArrayList)
            .flatMap(List::stream)
            .collect(Collectors.toList());
    // This happens when Gazelle is run in module mode, it wants to process the module level
    // directory, which would not
    // have any files. This is not an error, and should just be skipped. The IOException is caught
    // the next level up,
    // logged, and ignored.
    if (objectFiles.isEmpty()) {
      logger.debug("JavaTools: No files given to parse, skipping directory: {}", directory);
      throw new IOException("No files to process");
    }
    parseFileGatherDependencies(objectFiles);
  }

  public void parseClasses(List<? extends JavaFileObject> files) throws IOException {
    this.parseFileGatherDependencies(files);
  }

  private void parseFileGatherDependencies(Iterable<? extends JavaFileObject> compUnits)
      throws IOException {
    JavacTask task = (JavacTask) compiler.getTask(null, null, null, OPTIONS, null, compUnits);
    try {
      ClassScanner scanner = new ClassScanner();
      for (CompilationUnitTree compileUnitTree : task.parse()) {
        compileUnitTree.accept(scanner, null);
      }
    } catch (IOException ioException) {
      logger.error("JavaTools unable to read file(s)", ioException);
      throw ioException;
    } catch (Exception exception) {
      logger.error("JavaTools failed to parse {}, skipping file", compUnits, exception);
    }
  }

  class ClassScanner extends TreeScanner<Void, Void> {
    private CompilationUnitTree compileUnit;
    private String fileName;
    @Nullable private String currentClassName;

    @Override
    public Void visitCompilationUnit(CompilationUnitTree t, Void v) {
      compileUnit = t;
      fileName = Paths.get(compileUnit.getSourceFile().toUri().getPath()).getFileName().toString();
      currentClassName = null;
      return super.visitCompilationUnit(t, v);
    }

    @Override
    public Void visitPackage(PackageTree p, Void v) {
      logger.debug("JavaTools: Got package {} for class: {}", p.getPackageName(), fileName);
      packages.add(p.getPackageName().toString());
      return super.visitPackage(p, v);
    }

    /*
    This code needs to return "package.class" to match what the java/gazelle/private/java/imports.go
    expects. The imports code splits the string into package and class.
    There are two corner cases here:
    Wildcard imports: here we strip the ".*" off the end, matching what the import.go does. Essentially a package
    without a class.
    Static import: The assumption here is they are of the form "import package.class.name" so we strip the ".name" from
    the import to return just the "package.class" as expected.
     */
    @Override
    public Void visitImport(ImportTree i, Void v) {
      logger.debug(
          "JavaTools: found import static {}: {}", i.isStatic(), i.getQualifiedIdentifier());
      String name = i.getQualifiedIdentifier().toString();
      if (i.isStatic()) {
        String staticPackage = name.substring(0, name.lastIndexOf('.'));
        usedTypes.add(staticPackage);
      } else if (name.endsWith("*")) {
        String wildcardPackage = name.substring(0, name.lastIndexOf('.'));
        usedTypes.add(wildcardPackage);
      } else {
        usedTypes.add(name);
      }
      return super.visitImport(i, v);
    }

    @Override
    public Void visitClass(ClassTree t, Void v) {
      // Set class name for the top level classes only
      if (currentClassName == null) {
        currentClassName = t.getSimpleName().toString();
      }
      return super.visitClass(t, v);
    }

    @Override
    public Void visitMethod(com.sun.source.tree.MethodTree m, Void v) {
      boolean isVoidReturn = false;

      // Check the return type on the method.
      // void -> May be a main method
      // Identifier or Member Select -> May be a fully qualified name and needs to be included in
      // the types list
      if (m.getReturnType() != null
          && m.getReturnType().getKind() == Tree.Kind.PRIMITIVE_TYPE
          && ((PrimitiveTypeTree) m.getReturnType()).getPrimitiveTypeKind() == TypeKind.VOID) {
        isVoidReturn = true;
      } else if (m.getReturnType() != null) {
        checkFullyQualifiedType(m.getReturnType());
      }

      // Check to see if we have a main method
      if (m.getName().toString().equals("main")
          && m.getModifiers().getFlags().containsAll(Set.of(STATIC, PUBLIC))
          && isVoidReturn) {
        logger.debug("JavaTools: Found main method for {}", currentClassName);
        mainClasses.add(currentClassName);
      }

      // Check the parameters for the method
      // Identifier or Member Select -> May be a fully qualifyed type name
      // Parameterized Type -> The generics values may be qualified.
      // Arry -> The array may be a fully qul
      for (VariableTree param : m.getParameters()) {
        checkFullyQualifiedType(param.getType());
      }
      return super.visitMethod(m, v);
    }

    private void checkFullyQualifiedType(Tree identifier) {
      if (identifier.getKind() == Tree.Kind.IDENTIFIER
          || identifier.getKind() == Tree.Kind.MEMBER_SELECT) {
        String typeName = identifier.toString();
        if (typeName.contains(".")) {
          usedTypes.add(typeName);
        }
      } else if (identifier.getKind() == Tree.Kind.PARAMETERIZED_TYPE) {
        Tree baseType = ((ParameterizedTypeTree) identifier).getType();
        checkFullyQualifiedType(baseType);
        ((ParameterizedTypeTree) identifier)
            .getTypeArguments()
            .forEach(this::checkFullyQualifiedType);
      } else if (identifier.getKind() == Tree.Kind.ARRAY_TYPE) {
        Tree baseType = ((ArrayTypeTree) identifier).getType();
        checkFullyQualifiedType(baseType);
      }
    }
  }
}

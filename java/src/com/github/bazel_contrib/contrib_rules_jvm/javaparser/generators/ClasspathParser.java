package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static javax.lang.model.element.Modifier.PRIVATE;
import static javax.lang.model.element.Modifier.PUBLIC;
import static javax.lang.model.element.Modifier.STATIC;

import com.google.common.base.Joiner;
import com.google.common.base.Splitter;
import com.google.common.collect.ImmutableSet;
import com.google.common.collect.Lists;
import com.sun.source.tree.AnnotationTree;
import com.sun.source.tree.ArrayTypeTree;
import com.sun.source.tree.ClassTree;
import com.sun.source.tree.CompilationUnitTree;
import com.sun.source.tree.ExpressionTree;
import com.sun.source.tree.ImportTree;
import com.sun.source.tree.MemberSelectTree;
import com.sun.source.tree.MethodInvocationTree;
import com.sun.source.tree.NewClassTree;
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
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.Set;
import java.util.SortedMap;
import java.util.SortedSet;
import java.util.TreeMap;
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
  private final Set<String> usedPackagesWithoutSpecificTypes = new TreeSet<>();

  private final Set<String> exportedTypes = new TreeSet<>();
  private final Set<String> packages = new TreeSet<>();
  private final Set<String> mainClasses = new TreeSet<>();

  // Mapping from fully-qualified class-name to class-names of annotations on that class.
  // Annotations will be fully-qualified where that's known, and not where not known.
  final Map<String, PerClassData> perClassData = new TreeMap<>();

  // get the system java compiler instance
  private static final JavaCompiler compiler = ToolProvider.getSystemJavaCompiler();
  private static final List<String> OPTIONS = List.of("--release=" + Runtime.version().feature());

  public ClasspathParser() {
    // Doesn't need to do anything currently
  }

  static class PerClassData {
    public PerClassData() {
      this(new TreeSet<>(), new TreeMap<>());
    }

    @Override
    public String toString() {
      return "PerClassData{"
          + "annotations="
          + annotations
          + ", perMethodAnnotations="
          + perMethodAnnotations
          + '}';
    }

    public PerClassData(
        SortedSet<String> annotations, SortedMap<String, SortedSet<String>> perMethodAnnotations) {
      this.annotations = annotations;
      this.perMethodAnnotations = perMethodAnnotations;
    }

    final SortedSet<String> annotations;

    final SortedMap<String, SortedSet<String>> perMethodAnnotations;

    @Override
    public boolean equals(Object o) {
      if (this == o) return true;
      if (o == null || getClass() != o.getClass()) return false;
      PerClassData that = (PerClassData) o;
      return Objects.equals(annotations, that.annotations)
          && Objects.equals(perMethodAnnotations, that.perMethodAnnotations);
    }

    @Override
    public int hashCode() {
      return Objects.hash(annotations, perMethodAnnotations);
    }
  }

  public ImmutableSet<String> getUsedTypes() {
    return ImmutableSet.copyOf(usedTypes);
  }

  public ImmutableSet<String> getUsedPackagesWithoutSpecificTypes() {
    return ImmutableSet.copyOf(usedPackagesWithoutSpecificTypes);
  }

  public ImmutableSet<String> getExportedTypes() {
    return ImmutableSet.copyOf(exportedTypes);
  }

  public ImmutableSet<String> getPackages() {
    return ImmutableSet.copyOf(packages);
  }

  public ImmutableSet<String> getMainClasses() {
    return ImmutableSet.copyOf(mainClasses);
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
    @Nullable private String currentPackage;

    // Stack of possibly-nested contexts we may currently be in.
    // First element is the outer-most context (e.g. top-level class), last element is the inner-most context (e.g. inner class).
    // Currently tracks classes, so that we can know what outer and inner classes we may be in.
    private final Deque<Tree> stack = new ArrayDeque<>();

    @Nullable private Map<String, String> currentFileImports;

    void popOrThrow(Tree expected) {
      Tree popped = stack.removeLast();
      if (!expected.equals(popped)) {
        throw new IllegalStateException(String.format("Expected to pop %s but got %s", expected, popped));
      }
    }

    @Override
    public Void visitCompilationUnit(CompilationUnitTree t, Void v) {
      compileUnit = t;
      fileName = Paths.get(compileUnit.getSourceFile().toUri()).getFileName().toString();
      currentFileImports = new HashMap<>();

      return super.visitCompilationUnit(t, v);
    }

    @Override
    public Void visitPackage(PackageTree p, Void v) {
      logger.debug("JavaTools: Got package {} for class: {}", p.getPackageName(), fileName);
      packages.add(p.getPackageName().toString());
      currentPackage = p.getPackageName().toString();
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
      } else if (name.endsWith(".*")) {
        String wildcardPackage = name.substring(0, name.lastIndexOf('.'));
        usedPackagesWithoutSpecificTypes.add(wildcardPackage);
      } else {
        String[] parts = i.getQualifiedIdentifier().toString().split("\\.");
        currentFileImports.put(parts[parts.length - 1], i.getQualifiedIdentifier().toString());
        usedTypes.add(name);
      }
      return super.visitImport(i, v);
    }

    @Override
    public Void visitClass(ClassTree t, Void v) {
      stack.addLast(t);
      for (AnnotationTree annotation : t.getModifiers().getAnnotations()) {
        String annotationClassName = annotation.getAnnotationType().toString();
        String importedFullyQualified = currentFileImports.get(annotationClassName);
        String currentFullyQualifiedClass = currentFullyQualifiedClassName();
        if (importedFullyQualified != null) {
          noteAnnotatedClass(currentFullyQualifiedClass, importedFullyQualified);
        } else {
          noteAnnotatedClass(currentFullyQualifiedClass, annotationClassName);
        }
      }
      Void ret = super.visitClass(t, v);
      popOrThrow(t);
      return ret;
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
        Set<String> types = checkFullyQualifiedType(m.getReturnType());
        if (!m.getModifiers().getFlags().contains(PRIVATE)) {
          exportedTypes.addAll(types);
        }
      }

      handleAnnotations(m.getModifiers().getAnnotations());

      // Check to see if we have a main method
      if (m.getName().toString().equals("main")
          && m.getModifiers().getFlags().containsAll(Set.of(STATIC, PUBLIC))
          && isVoidReturn) {
        String currentClassName = currentNestedClassNameWithoutPackage();
        logger.debug("JavaTools: Found main method for {}", currentClassName);
        mainClasses.add(currentClassName);
      }

      // Check the parameters for the method
      for (VariableTree param : m.getParameters()) {
        checkFullyQualifiedType(param.getType());
        handleAnnotations(param.getModifiers().getAnnotations());
      }

      for (AnnotationTree annotation : m.getModifiers().getAnnotations()) {
        String annotationClassName = annotation.getAnnotationType().toString();
        String importedFullyQualified = currentFileImports.get(annotationClassName);
        String currentFullyQualifiedClass = currentFullyQualifiedClassName();
        if (importedFullyQualified != null) {
          noteAnnotatedMethod(
              currentFullyQualifiedClass, m.getName().toString(), importedFullyQualified);
        } else {
          noteAnnotatedMethod(
              currentFullyQualifiedClass, m.getName().toString(), annotationClassName);
        }
      }

      return super.visitMethod(m, v);
    }

    private void handleAnnotations(List<? extends AnnotationTree> annotations) {
      for (AnnotationTree annotation : annotations) {
        checkFullyQualifiedType(annotation.getAnnotationType());
      }
    }

    @Override
    public Void visitMethodInvocation(MethodInvocationTree node, Void v) {
      if (node.getMethodSelect() instanceof MemberSelectTree) {
        ExpressionTree container = ((MemberSelectTree) node.getMethodSelect()).getExpression();
        if (container instanceof MemberSelectTree) {
          MemberSelectTree containerMST = (MemberSelectTree) container;
          if (looksLikeClassName(containerMST.getIdentifier().toString())) {
            checkFullyQualifiedType(container);
          }
        }
      }
      return super.visitMethodInvocation(node, v);
    }

    private boolean looksLikeClassName(String identifier) {
      if (identifier.isEmpty()) {
        return false;
      }
      // Classes start with UpperCase.
      if (!Character.isUpperCase(identifier.charAt(0))) {
        return false;
      }
      // Single-char upper-case may well be a class-name.
      if (identifier.length() == 1) {
        return true;
      }
      // SNAKE_CASE is for constants not classes.
      if (identifier.chars().allMatch(c -> Character.isUpperCase(c) || c == '_')) {
        return false;
      }
      return true;
    }

    @Override
    public Void visitNewClass(NewClassTree node, Void v) {
      checkFullyQualifiedType(node.getIdentifier());
      return super.visitNewClass(node, v);
    }

    @Override
    public Void visitVariable(VariableTree node, Void unused) {
      if (node.getType() != null) {
        checkFullyQualifiedType(node.getType());
      }
      return super.visitVariable(node, unused);
    }

    @Nullable
    private Set<String> checkFullyQualifiedType(Tree identifier) {
      Set<String> types = new TreeSet<>();
      if (identifier.getKind() == Tree.Kind.IDENTIFIER
          || identifier.getKind() == Tree.Kind.MEMBER_SELECT) {
        String typeName = identifier.toString();
        List<String> components = Splitter.on(".").splitToList(typeName);
        if (currentFileImports.containsKey(components.get(0))) {
          String importedType = currentFileImports.get(components.get(0));
          usedTypes.add(importedType);
          types.add(importedType);
        } else if (components.size() > 1) {
          usedTypes.add(typeName);
          types.add(typeName);
        }
      } else if (identifier.getKind() == Tree.Kind.PARAMETERIZED_TYPE) {
        Tree baseType = ((ParameterizedTypeTree) identifier).getType();
        checkFullyQualifiedType(baseType);
        ((ParameterizedTypeTree) identifier)
            .getTypeArguments().stream()
                .flatMap(t -> checkFullyQualifiedType(t).stream())
                .forEach(types::add);
      } else if (identifier.getKind() == Tree.Kind.ARRAY_TYPE) {
        Tree baseType = ((ArrayTypeTree) identifier).getType();
        types.addAll(checkFullyQualifiedType(baseType));
      }
      return types;
    }

    @Nullable
    private String currentNestedClassNameWithoutPackage() {
      List<String> parts = new ArrayList<>();
      boolean sawClass = false;
      for (Tree tree : stack) {
        if (tree instanceof ClassTree) {
          sawClass = true;
          parts.add(((ClassTree) tree).getSimpleName().toString());
        }
      }
      if (!sawClass) {
        return null;
      }
      return Joiner.on('.').join(parts);
    }

    @Nullable
    private String currentFullyQualifiedClassName() {
      String nestedClassName = currentNestedClassNameWithoutPackage();
      if (nestedClassName == null) {
        return null;
      }
      List<String> parts = new ArrayList<>();
      if (currentPackage != null) {
        parts.add(currentPackage);
      }
      parts.add(nestedClassName);
      return Joiner.on('.').join(parts);
    }
  }

  private void noteAnnotatedClass(
      String annotatedFullyQualifiedClassName, String annotationFullyQualifiedClassName) {
    if (!perClassData.containsKey(annotatedFullyQualifiedClassName)) {
      perClassData.put(annotatedFullyQualifiedClassName, new PerClassData());
    }
    perClassData
        .get(annotatedFullyQualifiedClassName)
        .annotations
        .add(annotationFullyQualifiedClassName);
  }

  private void noteAnnotatedMethod(
      String annotatedFullyQualifiedClassName,
      String methodName,
      String annotationFullyQualifiedClassName) {
    if (!perClassData.containsKey(annotatedFullyQualifiedClassName)) {
      perClassData.put(annotatedFullyQualifiedClassName, new PerClassData());
    }
    PerClassData data = perClassData.get(annotatedFullyQualifiedClassName);
    if (!data.perMethodAnnotations.containsKey(methodName)) {
      data.perMethodAnnotations.put(methodName, new TreeSet<>());
    }
    data.perMethodAnnotations.get(methodName).add(annotationFullyQualifiedClassName);
  }
}

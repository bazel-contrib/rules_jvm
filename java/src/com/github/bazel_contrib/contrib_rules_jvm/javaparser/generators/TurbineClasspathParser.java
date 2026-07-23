package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators.ClassNames.isLikelyClassName;

import com.google.common.base.Joiner;
import com.google.turbine.diag.SourceFile;
import com.google.turbine.parse.Parser;
import com.google.turbine.tree.Tree;
import com.google.turbine.tree.TurbineModifier;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeSet;
import java.util.stream.Collectors;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Parses Java source files using Turbine's header-only parser instead of javac's full AST.
 *
 * <p>Turbine is stateless and single-pass, making it 50–60x faster than the javac pipeline for
 * large codebases. The trade-off: Turbine strips method bodies, so fully-qualified type references
 * used only inside methods (without a corresponding import) are not captured. In practice this is
 * rare, as style tools require explicit imports for every used type.
 */
public class TurbineClasspathParser {

  private static final Logger logger = LoggerFactory.getLogger(TurbineClasspathParser.class);

  public ParsedPackageData parseClasses(Path directory, List<String> files) throws IOException {
    ParsedPackageData data = new ParsedPackageData();
    for (String filename : files) {
      Path path = directory.resolve(filename);
      try {
        String source = Files.readString(path);
        data.merge(parseOneFile(path, source));
      } catch (IOException e) {
        logger.error("TurbineClasspathParser: cannot read {}", path, e);
      }
    }
    return data;
  }

  static ParsedPackageData parseOneFile(Path path, String source) {
    ParsedPackageData data = new ParsedPackageData();
    try {
      Tree.CompUnit unit = Parser.parse(new SourceFile(path.toString(), source));
      new Scanner(data).scan(unit);
    } catch (Throwable e) {
      // TurbineError extends Error, not Exception
      logger.error("TurbineClasspathParser: failed to parse {}", path, e);
    }
    return data;
  }

  static class Scanner {
    private final ParsedPackageData data;
    private String currentPackage = null;
    private Map<String, String> imports;
    private Set<String> locallyDefinedClassNames;
    private Set<String> typeParameterNames;
    private final Deque<String> classNameStack = new ArrayDeque<>();

    Scanner(ParsedPackageData data) {
      this.data = data;
    }

    void scan(Tree.CompUnit unit) {
      locallyDefinedClassNames = new TreeSet<>();
      collectLocalClassNames(unit.decls(), locallyDefinedClassNames);
      typeParameterNames = new TreeSet<>();
      imports = new HashMap<>();

      unit.pkg()
          .ifPresent(
              pkg -> {
                currentPackage = joinIdents(pkg.name());
                data.packages.add(currentPackage);
              });

      for (Tree.ImportDecl imp : unit.imports()) {
        String name = joinIdents(imp.type());
        if (imp.stat()) {
          if (imp.wild()) {
            // import static foo.Bar.* — add the class name (Bar) to usedTypes
            data.usedTypes.add(name);
          } else {
            // import static foo.Outer.MEMBER or foo.Outer.Inner
            int lastDot = name.lastIndexOf('.');
            String parentFqn = name.substring(0, lastDot);
            String lastSegment = name.substring(lastDot + 1);
            data.usedTypes.add(parentFqn);
            if (isLikelyClassName(lastSegment)) {
              imports.put(lastSegment, name);
              data.usedTypes.add(name);
            }
          }
        } else if (imp.wild()) {
          // import foo.bar.*
          data.usedPackagesWithoutSpecificTypes.add(name);
        } else {
          // import foo.bar.Baz
          String[] parts = name.split("\\.");
          imports.put(parts[parts.length - 1], name);
          data.usedTypes.add(name);
        }
      }

      for (Tree.TyDecl decl : unit.decls()) {
        processTypeDecl(decl);
      }
    }

    private void collectLocalClassNames(List<Tree.TyDecl> decls, Set<String> names) {
      for (Tree.TyDecl decl : decls) {
        String name = decl.name().value();
        if (!name.isEmpty()) {
          names.add(name);
        }
        for (Tree member : decl.members()) {
          if (member instanceof Tree.TyDecl) {
            collectLocalClassNames(List.of((Tree.TyDecl) member), names);
          }
        }
      }
    }

    private void processTypeDecl(Tree.TyDecl decl) {
      classNameStack.addLast(decl.name().value());

      for (Tree.TyParam tp : decl.typarams()) {
        typeParameterNames.add(tp.name().value());
        for (Tree bound : tp.bounds()) {
          resolveAndAddToUsedTypes(bound);
        }
      }

      decl.xtnds().ifPresent(this::resolveAndAddToUsedTypes);
      decl.impls().forEach(this::resolveAndAddToUsedTypes);

      String classFqn = currentFullyQualifiedClassName();
      for (Tree.Anno anno : decl.annos()) {
        String annoName = joinIdents(anno.name());
        TypeNameResolver.resolve(annoName, imports, currentPackage, excludedNames())
            .ifPresent(data.usedTypes::add);
        noteAnnotatedClass(classFqn, resolveAnnotationName(annoName));
      }

      for (Tree member : decl.members()) {
        if (member instanceof Tree.MethDecl) {
          processMethodDecl((Tree.MethDecl) member, classFqn);
        } else if (member instanceof Tree.VarDecl) {
          processFieldDecl((Tree.VarDecl) member, classFqn);
        } else if (member instanceof Tree.TyDecl) {
          processTypeDecl((Tree.TyDecl) member);
        }
      }

      classNameStack.removeLast();
    }

    private void processMethodDecl(Tree.MethDecl meth, String classFqn) {
      for (Tree.TyParam tp : meth.typarams()) {
        typeParameterNames.add(tp.name().value());
        for (Tree bound : tp.bounds()) {
          resolveAndAddToUsedTypes(bound);
        }
      }

      boolean isPrivate = meth.mods().contains(TurbineModifier.PRIVATE);
      boolean isVoidReturn = false;

      if (meth.ret().isPresent()) {
        Tree ret = meth.ret().get();
        if (ret.kind() == Tree.Kind.VOID_TY) {
          isVoidReturn = true;
        } else {
          Set<String> types = resolveTree(ret);
          if (!isPrivate) {
            data.exportedTypes.addAll(types);
          }
        }
      }

      for (Tree.ClassTy exnty : meth.exntys()) {
        resolveAndAddToUsedTypes(exnty);
      }

      for (Tree.Anno anno : meth.annos()) {
        String annoName = joinIdents(anno.name());
        TypeNameResolver.resolve(annoName, imports, currentPackage, excludedNames())
            .ifPresent(data.usedTypes::add);
        noteAnnotatedMethod(classFqn, meth.name().value(), resolveAnnotationName(annoName));
      }

      for (Tree.VarDecl param : meth.params()) {
        Set<String> paramTypes = resolveTree(param.ty());
        if (!isPrivate) {
          data.exportedTypes.addAll(paramTypes);
        }
        for (Tree.Anno anno : param.annos()) {
          TypeNameResolver.resolve(
                  joinIdents(anno.name()), imports, currentPackage, excludedNames())
              .ifPresent(data.usedTypes::add);
        }
      }

      if (isVoidReturn
          && meth.name().value().equals("main")
          && meth.mods().contains(TurbineModifier.STATIC)
          && meth.mods().contains(TurbineModifier.PUBLIC)) {
        String className = currentNestedClassNameWithoutPackage();
        if (className != null) {
          data.mainClasses.add(className);
        }
      }
    }

    private void processFieldDecl(Tree.VarDecl field, String classFqn) {
      boolean isPrivate = field.mods().contains(TurbineModifier.PRIVATE);
      Set<String> fieldTypes = resolveTree(field.ty());
      if (!isPrivate) {
        data.exportedTypes.addAll(fieldTypes);
      }

      for (Tree.Anno anno : field.annos()) {
        String annoName = joinIdents(anno.name());
        TypeNameResolver.resolve(annoName, imports, currentPackage, excludedNames())
            .ifPresent(data.usedTypes::add);
        noteAnnotatedField(classFqn, field.name().value(), resolveAnnotationName(annoName));
      }
    }

    private Set<String> resolveTree(Tree tree) {
      Set<String> result = new TreeSet<>();
      if (tree instanceof Tree.ClassTy) {
        result.addAll(resolveClassTy((Tree.ClassTy) tree));
      } else if (tree instanceof Tree.ArrTy) {
        result.addAll(resolveTree(((Tree.ArrTy) tree).elem()));
      } else if (tree instanceof Tree.WildTy) {
        Tree.WildTy wild = (Tree.WildTy) tree;
        wild.upper().ifPresent(t -> result.addAll(resolveTree(t)));
        wild.lower().ifPresent(t -> result.addAll(resolveTree(t)));
      }
      return result;
    }

    private Set<String> resolveClassTy(Tree.ClassTy ty) {
      Set<String> result = new TreeSet<>();
      TypeNameResolver.resolve(classTypeName(ty), imports, currentPackage, excludedNames())
          .ifPresent(
              resolved -> {
                result.add(resolved);
                data.usedTypes.add(resolved);
              });
      for (Tree tyarg : ty.tyargs()) {
        result.addAll(resolveTree(tyarg));
      }
      return result;
    }

    private void resolveAndAddToUsedTypes(Tree tree) {
      resolveTree(tree).forEach(data.usedTypes::add);
    }

    private String classTypeName(Tree.ClassTy ty) {
      if (ty.base().isEmpty()) {
        return ty.name().value();
      }
      return classTypeName(ty.base().get()) + "." + ty.name().value();
    }

    private String resolveAnnotationName(String annoName) {
      int firstDot = annoName.indexOf('.');
      String firstSegment = firstDot == -1 ? annoName : annoName.substring(0, firstDot);
      String imported = imports.get(firstSegment);
      if (imported != null) {
        return imported;
      }
      return annoName;
    }

    private Set<String> excludedNames() {
      Set<String> excluded = new TreeSet<>(ClassNames.JAVA_LANG_TYPES);
      excluded.addAll(locallyDefinedClassNames);
      excluded.addAll(typeParameterNames);
      return excluded;
    }

    private String joinIdents(List<Tree.Ident> idents) {
      return idents.stream().map(Tree.Ident::value).collect(Collectors.joining("."));
    }

    private String currentNestedClassNameWithoutPackage() {
      if (classNameStack.isEmpty()) {
        return null;
      }
      List<String> parts = new ArrayList<>(classNameStack);
      return Joiner.on('.').join(parts);
    }

    private String currentFullyQualifiedClassName() {
      String nested = currentNestedClassNameWithoutPackage();
      if (nested == null) {
        return null;
      }
      if (currentPackage == null || currentPackage.isEmpty()) {
        return nested;
      }
      return currentPackage + "." + nested;
    }

    private void noteAnnotatedClass(String classFqn, String annoFqn) {
      if (classFqn == null) {
        return;
      }
      data.perClassData.computeIfAbsent(classFqn, k -> new PerClassData()).annotations.add(annoFqn);
    }

    private void noteAnnotatedMethod(String classFqn, String methodName, String annoFqn) {
      if (classFqn == null) {
        return;
      }
      data.perClassData
          .computeIfAbsent(classFqn, k -> new PerClassData())
          .perMethodAnnotations
          .computeIfAbsent(methodName, k -> new TreeSet<>())
          .add(annoFqn);
    }

    private void noteAnnotatedField(String classFqn, String fieldName, String annoFqn) {
      if (classFqn == null) {
        return;
      }
      data.perClassData
          .computeIfAbsent(classFqn, k -> new PerClassData())
          .perFieldAnnotations
          .computeIfAbsent(fieldName, k -> new TreeSet<>())
          .add(annoFqn);
    }
  }
}

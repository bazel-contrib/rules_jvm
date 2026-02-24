package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.openapi.vfs.VirtualFileManager;
import com.intellij.psi.PsiManager;
import com.intellij.psi.tree.IElementType;
import java.nio.file.Path;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.Stack;
import java.util.TreeSet;
import java.util.stream.Collectors;
import org.jetbrains.kotlin.cli.jvm.compiler.EnvironmentConfigFiles;
import org.jetbrains.kotlin.cli.jvm.compiler.KotlinCoreEnvironment;
import org.jetbrains.kotlin.config.CommonConfigurationKeys;
import org.jetbrains.kotlin.config.CompilerConfiguration;
import org.jetbrains.kotlin.lexer.KtTokens;
import org.jetbrains.kotlin.name.FqName;
import org.jetbrains.kotlin.name.FqNamesUtilKt;
import org.jetbrains.kotlin.name.Name;
import org.jetbrains.kotlin.name.NameUtils;
import org.jetbrains.kotlin.psi.KtAnnotated;
import org.jetbrains.kotlin.psi.KtBinaryExpression;
import org.jetbrains.kotlin.psi.KtCallExpression;
import org.jetbrains.kotlin.psi.KtClass;
import org.jetbrains.kotlin.psi.KtClassBody;
import org.jetbrains.kotlin.psi.KtDestructuringDeclaration;
import org.jetbrains.kotlin.psi.KtDestructuringDeclarationEntry;
import org.jetbrains.kotlin.psi.KtDotQualifiedExpression;
import org.jetbrains.kotlin.psi.KtElement;
import org.jetbrains.kotlin.psi.KtExpression;
import org.jetbrains.kotlin.psi.KtFile;
import org.jetbrains.kotlin.psi.KtImportDirective;
import org.jetbrains.kotlin.psi.KtModifierListOwner;
import org.jetbrains.kotlin.psi.KtNamedFunction;
import org.jetbrains.kotlin.psi.KtNullableType;
import org.jetbrains.kotlin.psi.KtObjectDeclaration;
import org.jetbrains.kotlin.psi.KtPackageDirective;
import org.jetbrains.kotlin.psi.KtParameter;
import org.jetbrains.kotlin.psi.KtProperty;
import org.jetbrains.kotlin.psi.KtQualifiedExpression;
import org.jetbrains.kotlin.psi.KtReferenceExpression;
import org.jetbrains.kotlin.psi.KtSafeQualifiedExpression;
import org.jetbrains.kotlin.psi.KtSimpleNameExpression;
import org.jetbrains.kotlin.psi.KtTreeVisitorVoid;
import org.jetbrains.kotlin.psi.KtTypeElement;
import org.jetbrains.kotlin.psi.KtTypeReference;
import org.jetbrains.kotlin.psi.KtUnaryExpression;
import org.jetbrains.kotlin.psi.KtUserType;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class KtParser {
  private static final Logger logger = LoggerFactory.getLogger(GrpcServer.class);

  private final CompilerConfiguration compilerConf = createCompilerConfiguration();
  private final KotlinCoreEnvironment env =
      KotlinCoreEnvironment.createForProduction(
          Disposer.newDisposable(), compilerConf, EnvironmentConfigFiles.JVM_CONFIG_FILES);
  private final VirtualFileManager vfm = VirtualFileManager.getInstance();
  private final PsiManager psiManager = PsiManager.getInstance(env.getProject());

  public ParsedPackageData parseClasses(List<Path> files) {
    KtFileVisitor visitor = new KtFileVisitor();
    List<VirtualFile> virtualFiles =
        files.stream().map(f -> vfm.findFileByNioPath(f)).collect(Collectors.toUnmodifiableList());

    for (VirtualFile virtualFile : virtualFiles) {
      if (virtualFile == null) {
        throw new IllegalArgumentException("File not found: " + files.get(0));
      }
      KtFile ktFile = (KtFile) psiManager.findFile(virtualFile);

      logger.debug("ktFile: {}", ktFile);
      logger.debug("name: {}", ktFile.getName());
      logger.debug("script: {}", ktFile.getScript());
      logger.debug("declarations: {}", ktFile.getDeclarations());
      logger.debug(
          "classes: {}",
          Arrays.stream(ktFile.getClasses())
              .map(Object::toString)
              .collect(Collectors.joining(", ")));
      logger.debug("import directives: {}", ktFile.getImportDirectives());
      logger.debug("import list: {}", ktFile.getImportList());

      ktFile.accept(visitor);
    }

    return visitor.packageData;
  }

  private static CompilerConfiguration createCompilerConfiguration() {
    CompilerConfiguration conf = new CompilerConfiguration();
    conf.put(CommonConfigurationKeys.MODULE_NAME, "bazel-module");

    return conf;
  }

  public static class KtFileVisitor extends KtTreeVisitorVoid {
    final ParsedPackageData packageData = new ParsedPackageData();

    private Stack<Visibility> visibilityStack = new Stack<>();
    private HashMap<String, FqName> fqImportByNameOrAlias = new HashMap<>();

    // Track inline functions and their dependencies
    private String currentInlineFunction = null;
    private Set<String> currentInlineFunctionDeps = null;

    // Track extension functions and their dependencies
    private String currentExtensionFunction = null;
    private String currentExtensionReceiverType = null;
    private Set<String> currentExtensionFunctionDeps = null;

    // Track property delegates and their dependencies
    private boolean currentlyInPropertyDelegate = false;
    private Set<String> currentPropertyDelegateDeps = null;

    // Track destructuring declarations and their dependencies
    private Set<String> detectedDestructuringDeclarations = new TreeSet<>();

    // Track componentN() functions and their dependencies for destructuring analysis
    private Map<String, Set<String>> componentFunctionDeps = new HashMap<>();

    // Track when we're inside a componentN() function to detect its dependencies
    private String currentComponentFunction = null;
    private Set<String> currentComponentFunctionDeps = null;

    @Override
    public void visitPackageDirective(KtPackageDirective packageDirective) {
      packageData.packages.add(packageDirective.getQualifiedName());
      super.visitPackageDirective(packageDirective);
    }

    @Override
    public void visitKtFile(KtFile file) {
      if (file.hasTopLevelCallables()) {
        FqName filePackage = file.getPackageFqName();
        String outerClassName = NameUtils.getScriptNameForFile(file.getName()) + "Kt";
        FqName outerClassFqName = filePackage.child(Name.identifier(outerClassName));
        packageData.perClassData.put(outerClassFqName.toString(), new PerClassData());
      }
      super.visitKtFile(file);

      // Log AST-based usage statistics
      logger.debug(
          "AST: Destructuring declarations detected: " + detectedDestructuringDeclarations.size());
    }

    @Override
    public void visitImportDirective(KtImportDirective importDirective) {
      FqName importName = importDirective.getImportedFqName();
      boolean foundClass = false;
      int segmentCount = 0;
      List<Name> pathSegments = importName.pathSegments();
      for (Name importPart : pathSegments) {
        segmentCount++;
        // If there is a PascalCase component, assume it's a class.
        if (isLikelyClassName(importPart.asString())) {
          foundClass = true;
          break;
        }
      }
      if (foundClass) {
        FqName className = importName;
        if (!isLikelyClassName(importName.shortName().asString())) {
          // If we're directly importing a function from a parent class or object, use the parent.
          className = importName.parent();
        }
        String localName = importDirective.getAliasName();
        if (localName == null) {
          localName = className.shortName().toString();
        }
        packageData.usedTypes.add(className.toString());
        fqImportByNameOrAlias.put(localName, className);
      } else {
        if (importDirective.isAllUnder()) {
          // If it's a wildcard import with no obvious class name, assume it's a package.
          packageData.usedPackagesWithoutSpecificTypes.add(importName.asString());
        } else {
          // If it's not a wildcard import and lacks an obvious class name, assume it's a function
          // in a package.
          packageData.usedPackagesWithoutSpecificTypes.add(importName.parent().asString());
        }
      }
      super.visitImportDirective(importDirective);
    }

    @Override
    public void visitClass(KtClass clazz) {
      pushState(clazz);
      if (clazz.isLocal() || !isVisible()) {
        super.visitClass(clazz);
        popState(clazz);
        return;
      }
      packageData.perClassData.put(clazz.getFqName().toString(), new PerClassData());
      super.visitClass(clazz);
      popState(clazz);
    }

    @Override
    public void visitObjectDeclaration(KtObjectDeclaration object) {
      pushState(object);
      if (object.isLocal() || !isVisible()) {
        super.visitObjectDeclaration(object);
        popState(object);
        return;
      }

      packageData.perClassData.put(object.getFqName().toString(), new PerClassData());

      super.visitObjectDeclaration(object);
      popState(object);
    }

    @Override
    public void visitProperty(KtProperty property) {
      pushState(property);
      if (property.isLocal() || !isVisible()) {
        super.visitProperty(property);
        popState(property);
        return;
      }

      KtTypeReference typeReference = property.getTypeReference();
      if (typeReference != null) {
        addExportedTypeIfNeeded(typeReference);
      }

      // Check if this property has a delegate (uses 'by' keyword)
      if (property.hasDelegate()) {
        FqName propertyFqName = getPropertyFqName(property);
        currentlyInPropertyDelegate = true;
        currentPropertyDelegateDeps = new TreeSet<>();

        // Get delegate expression to determine delegate type
        KtExpression delegateExpression = property.getDelegateExpression();
        String delegateType = "unknown";
        if (delegateExpression != null) {
          delegateType = getDelegateType(delegateExpression);
          logger.debug(
              "Found property delegate: "
                  + propertyFqName
                  + " with delegate type: "
                  + delegateType);
        }
      }

      super.visitProperty(property);

      // If this was a property delegate, add its dependencies to exported types
      if (property.hasDelegate() && currentlyInPropertyDelegate) {
        packageData.exportedTypes.addAll(currentPropertyDelegateDeps);
        logger.debug(
            "Property delegate "
                + getPropertyFqName(property)
                + " added "
                + currentPropertyDelegateDeps.size()
                + " exported types (from property delegate): "
                + currentPropertyDelegateDeps);
        currentlyInPropertyDelegate = false;
        currentPropertyDelegateDeps = null;
      }

      popState(property);
    }

    @Override
    public void visitNamedFunction(KtNamedFunction function) {
      pushState(function);
      if (function.isLocal() || !isVisible()) {
        super.visitNamedFunction(function);
        popState(function);
        return;
      }

      // Check if this is a componentN() function for destructuring
      boolean isComponentFunction =
          function.getName() != null
              && function.getName().startsWith("component")
              && function.hasModifier(KtTokens.OPERATOR_KEYWORD);
      if (isComponentFunction) {
        // Start tracking dependencies for this componentN() function
        FqName functionFqName = getFunctionFqName(function);
        currentComponentFunction = functionFqName.toString();
        currentComponentFunctionDeps = new TreeSet<>();
        logger.debug("Found componentN() function: " + currentComponentFunction);
      }

      // Check if this is an inline function
      boolean isInline = function.hasModifier(KtTokens.INLINE_KEYWORD);
      logger.debug(
          "Checking function: "
              + function.getName()
              + ", isInline: "
              + isInline
              + ", modifiers: "
              + function.getModifierList());
      if (isInline) {
        // Start tracking dependencies for this inline function
        FqName functionFqName = getFunctionFqName(function);
        currentInlineFunction = functionFqName.toString();
        currentInlineFunctionDeps = new TreeSet<>();
        logger.debug("Found inline function: " + currentInlineFunction);
      }
      // Check if this is an extension function
      boolean isExtension = function.getReceiverTypeReference() != null;
      logger.debug("Checking function: " + function.getName() + ", isExtension: " + isExtension);
      if (isExtension) {
        // Start tracking dependencies for this extension function
        FqName functionFqName = getFunctionFqName(function);
        currentExtensionFunction = functionFqName.toString();
        currentExtensionFunctionDeps = new TreeSet<>();

        // Get receiver type
        KtTypeReference receiverTypeRef = function.getReceiverTypeReference();
        if (receiverTypeRef != null) {
          currentExtensionReceiverType = getTypeString(receiverTypeRef);
          logger.debug(
              "Found extension function: "
                  + currentExtensionFunction
                  + " extending "
                  + currentExtensionReceiverType);
        } else {
          currentExtensionReceiverType = "unknown";
        }
      }

      if (hasMainFunctionShape(function)) {
        if (function.isTopLevel()) {
          FqName outerClassFqName = javaClassNameForKtFile(function.getContainingKtFile());
          String outerClassName = outerClassFqName.shortName().asString();
          packageData.mainClasses.add(outerClassName);
        } else if (function.getParent().getParent() instanceof KtObjectDeclaration) {
          // The parent is a class/object body, then the next parent should be a class or object.
          KtObjectDeclaration object = (KtObjectDeclaration) function.getParent().getParent();
          FqName relativeFqName =
              packageRelativeName(object.getFqName(), object.getContainingKtFile());
          if (isJvmStatic(function)) {
            packageData.mainClasses.add(relativeFqName.parent().toString());
          } else {
            packageData.mainClasses.add(relativeFqName.asString());
          }
        }
      }

      KtTypeReference returnType = function.getTypeReference();
      if (returnType != null) {
        addExportedTypeIfNeeded(returnType);
      }

      // Visit the function body to collect dependencies
      super.visitNamedFunction(function);

      // If this was an inline function, add its dependencies to exported types
      if (isInline) {
        packageData.exportedTypes.addAll(currentInlineFunctionDeps);
        logger.debug(
            "Inline function "
                + currentInlineFunction
                + " added "
                + currentInlineFunctionDeps.size()
                + " exported types (from inline function): "
                + currentInlineFunctionDeps);
        currentInlineFunction = null;
        currentInlineFunctionDeps = null;
      }
      // If this was an extension function, add its dependencies to exported types
      if (isExtension) {
        packageData.exportedTypes.addAll(currentExtensionFunctionDeps);
        logger.debug(
            "Extension function "
                + currentExtensionFunction
                + " extending "
                + currentExtensionReceiverType
                + " added "
                + currentExtensionFunctionDeps.size()
                + " exported types (from extension function): "
                + currentExtensionFunctionDeps);
        currentExtensionFunction = null;
        currentExtensionReceiverType = null;
        currentExtensionFunctionDeps = null;
      }

      // If this was a componentN() function, save its dependencies
      if (isComponentFunction) {
        componentFunctionDeps.put(currentComponentFunction, currentComponentFunctionDeps);
        // Also add these dependencies to the exported types since they'll be needed for
        // destructuring
        packageData.exportedTypes.addAll(currentComponentFunctionDeps);
        logger.debug(
            "ComponentN function "
                + currentComponentFunction
                + " uses: "
                + currentComponentFunctionDeps);
        currentComponentFunction = null;
        currentComponentFunctionDeps = null;
      }

      popState(function);
    }

    @Override
    public void visitQualifiedExpression(KtQualifiedExpression expression) {
      logger.debug("Qualified expression: " + expression.getText());

      super.visitQualifiedExpression(expression);
    }

    @Override
    public void visitTypeReference(KtTypeReference reference) {
      logger.debug("Type reference: " + reference.getText());

      super.visitTypeReference(reference);
    }

    @Override
    public void visitReferenceExpression(KtReferenceExpression reference) {
      logger.debug("Reference expression: " + reference.getText());

      super.visitReferenceExpression(reference);
    }

    @Override
    public void visitSimpleNameExpression(KtSimpleNameExpression expression) {
      logger.debug(
          "Simple name expression: {}, Referenced name: {}",
          expression.getText(),
          expression.getReferencedName());

      // If we're inside an inline function, track class usage
      if (currentInlineFunction != null) {
        String referencedName = expression.getReferencedName();

        // Check if this is a class reference (constructor call or type reference)
        // This includes both PascalCase class names and constructor calls
        if (isLikelyClassName(referencedName)) {
          // Try to resolve to fully qualified name
          if (fqImportByNameOrAlias.containsKey(referencedName)) {
            String fqName = fqImportByNameOrAlias.get(referencedName).toString();
            currentInlineFunctionDeps.add(fqName);
            logger.debug("Inline function " + currentInlineFunction + " uses class: " + fqName);
          } else if (referencedName.contains(".")) {
            // Already fully qualified
            currentInlineFunctionDeps.add(referencedName);
            logger.debug(
                "Inline function " + currentInlineFunction + " uses class: " + referencedName);
          } else {
            // For unresolved class names, add them as potential dependencies
            // This handles cases where imports might not be fully resolved
            logger.debug(
                "Inline function "
                    + currentInlineFunction
                    + " uses unresolved class: "
                    + referencedName);
          }
        }
      }

      // If we're inside an extension function, track class usage
      if (currentExtensionFunction != null) {
        String referencedName = expression.getReferencedName();

        // Check if this is a class reference (constructor call or type reference)
        if (isLikelyClassName(referencedName)) {
          // Try to resolve to fully qualified name
          if (fqImportByNameOrAlias.containsKey(referencedName)) {
            String fqName = fqImportByNameOrAlias.get(referencedName).toString();
            currentExtensionFunctionDeps.add(fqName);
            logger.debug(
                "Extension function " + currentExtensionFunction + " uses class: " + fqName);
          } else if (referencedName.contains(".")) {
            // Already fully qualified
            currentExtensionFunctionDeps.add(referencedName);
            logger.debug(
                "Extension function "
                    + currentExtensionFunction
                    + " uses class: "
                    + referencedName);
          } else {
            logger.debug(
                "Extension function "
                    + currentExtensionFunction
                    + " uses unresolved class: "
                    + referencedName);
          }
        }
      }

      // If we're inside a property delegate, track class usage
      if (currentlyInPropertyDelegate) {
        String referencedName = expression.getReferencedName();

        // Check if this is a class reference (constructor call or type reference)
        if (isLikelyClassName(referencedName)) {
          // Try to resolve to fully qualified name
          if (fqImportByNameOrAlias.containsKey(referencedName)) {
            String fqName = fqImportByNameOrAlias.get(referencedName).toString();
            currentPropertyDelegateDeps.add(fqName);
            logger.debug("Property delegate uses class: " + fqName);
          } else if (referencedName.contains(".")) {
            // Already fully qualified
            currentPropertyDelegateDeps.add(referencedName);
            logger.debug("Property delegate uses class: " + referencedName);
          } else {
            logger.debug("Property delegate uses unresolved class: " + referencedName);
          }
        }
      }

      // If we're inside a componentN() function, track class usage
      if (currentComponentFunction != null) {
        String referencedName = expression.getReferencedName();

        // Check if this is a class reference (constructor call or type reference)
        if (isLikelyClassName(referencedName)) {
          // Try to resolve to fully qualified name
          if (fqImportByNameOrAlias.containsKey(referencedName)) {
            String fqName = fqImportByNameOrAlias.get(referencedName).toString();
            currentComponentFunctionDeps.add(fqName);
            logger.debug(
                "ComponentN function " + currentComponentFunction + " uses class: " + fqName);
          } else if (referencedName.contains(".")) {
            // Already fully qualified
            currentComponentFunctionDeps.add(referencedName);
            logger.debug(
                "ComponentN function "
                    + currentComponentFunction
                    + " uses class: "
                    + referencedName);
          } else {
            logger.debug(
                "ComponentN function "
                    + currentComponentFunction
                    + " uses unresolved class: "
                    + referencedName);
          }
        }
      }

      super.visitSimpleNameExpression(expression);
    }

    @Override
    public void visitCallExpression(KtCallExpression expression) {
      logger.debug("AST: Call expression: " + expression.getText());

      KtExpression calleeExpression = expression.getCalleeExpression();
      if (calleeExpression != null) {
        String functionName = calleeExpression.getText();
        logger.debug("AST: Function call detected: " + functionName);

        // Check if this is a call to a known inline function
        checkInlineFunctionUsage(functionName);
      }

      super.visitCallExpression(expression);
    }

    @Override
    public void visitBinaryExpression(KtBinaryExpression expression) {
      logger.debug("AST: Binary expression: " + expression.getText());

      IElementType operationToken = expression.getOperationToken();
      if (operationToken != null) {
        String operator = operationToken.toString();
        logger.debug("AST: Operator usage detected: " + operator);

        checkExtensionOperatorUsage(operator, expression);
      }

      super.visitBinaryExpression(expression);
    }

    @Override
    public void visitUnaryExpression(KtUnaryExpression expression) {
      logger.debug("AST: Unary expression: " + expression.getText());

      IElementType operationToken = expression.getOperationToken();
      if (operationToken != null) {
        String operator = operationToken.toString();
        logger.debug("AST: Unary operator usage detected: " + operator);

        checkExtensionOperatorUsage(operator, expression);
      }

      super.visitUnaryExpression(expression);
    }

    @Override
    public void visitDotQualifiedExpression(KtDotQualifiedExpression expression) {
      logger.debug("AST: Dot qualified expression: " + expression.getText());

      KtExpression selectorExpression = expression.getSelectorExpression();
      if (selectorExpression instanceof KtCallExpression) {
        KtCallExpression callExpr = (KtCallExpression) selectorExpression;
        KtExpression receiverExpression = expression.getReceiverExpression();
        if (receiverExpression != null && callExpr.getCalleeExpression() != null) {
          String receiverType = getSimpleExpressionType(receiverExpression);
          String functionName = callExpr.getCalleeExpression().getText();

          checkExtensionFunctionCall(receiverType, functionName);

          // Detect FQN constructor call: com.example.ClassName(args)
          // Mirrors ClasspathParser.visitNewClass → checkFullyQualifiedType
          if (isLikelyClassName(functionName) && receiverType.contains(".")) {
            String fqClassName = receiverType + "." + functionName;
            packageData.usedTypes.add(fqClassName);
          }
        }
      }

      // Detect FQN class reference: com.example.ClassName (as selector of a DQE)
      // Mirrors ClasspathParser.visitMethodInvocation + looksLikeClassName
      if (selectorExpression instanceof KtSimpleNameExpression) {
        String selectorName = ((KtSimpleNameExpression) selectorExpression).getReferencedName();
        if (isLikelyClassName(selectorName)) {
          KtExpression receiverExpr = expression.getReceiverExpression();
          if (receiverExpr != null) {
            String receiverText = receiverExpr.getText();
            if (receiverText.contains(".")) {
              String fqClassName = receiverText + "." + selectorName;
              packageData.usedTypes.add(fqClassName);
            }
          }
        }
      }

      super.visitDotQualifiedExpression(expression);
    }

    @Override
    public void visitSafeQualifiedExpression(KtSafeQualifiedExpression expression) {
      logger.debug("AST: Safe qualified expression: " + expression.getText());

      KtExpression selectorExpression = expression.getSelectorExpression();
      if (selectorExpression instanceof KtCallExpression) {
        KtCallExpression callExpr = (KtCallExpression) selectorExpression;
        KtExpression receiverExpression = expression.getReceiverExpression();
        if (receiverExpression != null && callExpr.getCalleeExpression() != null) {
          String receiverType = getSimpleExpressionType(receiverExpression);
          String functionName = callExpr.getCalleeExpression().getText();

          checkExtensionFunctionCall(receiverType, functionName);
        }
      }

      super.visitSafeQualifiedExpression(expression);
    }

    @Override
    public void visitDestructuringDeclaration(KtDestructuringDeclaration declaration) {
      logger.debug("AST: Destructuring declaration: " + declaration.getText());

      String destructuringText = declaration.getText();
      detectedDestructuringDeclarations.add(destructuringText);

      // Get the initializer expression (the right side of the assignment)
      KtExpression initializer = declaration.getInitializer();
      if (initializer != null) {
        String initializerType = getSimpleExpressionType(initializer);
        logger.debug("AST: Destructuring initializer type: " + initializerType);

        // Get the number of components being destructured
        List<KtDestructuringDeclarationEntry> entries = declaration.getEntries();
        int componentCount = entries.size();

        logger.debug(
            "AST: Destructuring " + componentCount + " components from " + initializerType);

        // Check if this destructuring uses custom componentN() functions that might have
        // dependencies
        checkDestructuringDependencies(initializerType, componentCount, declaration);
      }

      super.visitDestructuringDeclaration(declaration);
    }

    /** Check if destructuring uses custom componentN() functions with external dependencies. */
    private void checkDestructuringDependencies(
        String initializerType, int componentCount, KtDestructuringDeclaration declaration) {
      logger.debug(
          "AST: Checking destructuring dependencies for "
              + initializerType
              + " with "
              + componentCount
              + " components");

      // Skip built-in types that have automatic componentN() functions
      if (isBuiltInDataType(initializerType)) {
        logger.debug("AST: Skipping built-in data type: " + initializerType);
        return;
      }

      // Try to resolve the initializer type to see if it's a custom class
      String resolvedType = resolveTypeToFqName(initializerType);
      if (resolvedType != null && !resolvedType.startsWith("kotlin.")) {
        // This appears to be a custom class - check if it defines componentN() functions
        logger.debug("AST: Custom type detected for destructuring: " + resolvedType);

        // For destructuring, we don't need to match specific componentN() functions
        // The componentN() functions have already been processed and their dependencies
        // added to exportedTypes when we visited them. The destructuring just triggers
        // the need for those dependencies to be available.

        // Since we're in the same package as the componentN() functions, and those
        // functions have already added their dependencies to exportedTypes, we don't
        // need to do anything additional here. The dependencies are already captured.

        logger.debug(
            "AST: Destructuring of "
                + resolvedType
                + " will use componentN() dependencies already captured");
      }
    }

    /**
     * Check if a type is a built-in data type that automatically provides componentN() functions.
     */
    private boolean isBuiltInDataType(String typeName) {
      // Built-in types that automatically provide componentN() functions
      return typeName.equals("Pair")
          || typeName.equals("Triple")
          || typeName.startsWith("kotlin.Pair")
          || typeName.startsWith("kotlin.Triple")
          ||
          // Data classes automatically generate componentN() functions,
          // but we can't easily detect if a class is a data class from just the type name
          // So we'll be conservative and only skip obviously built-in types
          false;
    }

    /** Resolve a type name to its fully qualified name using imports. */
    private String resolveTypeToFqName(String typeName) {
      // Try to resolve using imports
      if (fqImportByNameOrAlias.containsKey(typeName)) {
        return fqImportByNameOrAlias.get(typeName).toString();
      }

      // If it's already fully qualified, return as-is
      if (typeName.contains(".")) {
        return typeName;
      }

      // For unresolved types, return null
      return null;
    }

    /** Get statistics about destructuring declarations detected during parsing. */
    public int getDestructuringDeclarationCount() {
      return detectedDestructuringDeclarations.size();
    }

    /** Check if a function call is to a known inline function and track its usage. */
    private void checkInlineFunctionUsage(String functionName) {
      // This method is kept for potential future use but currently does nothing
      // since inline functions are now handled through implicit deps
      logger.debug("AST: Function call detected: " + functionName);
    }

    /** Check if an extension function is called on a specific receiver type. */
    private void checkExtensionFunctionCall(String receiverType, String functionName) {
      // This method is kept for potential future use but currently does nothing
      // since extension functions are now handled through implicit deps
      logger.debug("AST: Checking extension function call: " + receiverType + "." + functionName);
    }

    /** Check if an operator usage corresponds to an extension operator. */
    private void checkExtensionOperatorUsage(String operator, KtExpression expression) {
      // This method is kept for potential future use but currently does nothing
      // since extension operators are now handled through implicit deps
      logger.debug("AST: Checking extension operator usage: " + operator);
    }

    /** Map operator tokens to their corresponding function names. */
    private String mapOperatorToFunctionName(String operator) {
      switch (operator) {
        case "PLUS":
          return "plus";
        case "MINUS":
          return "minus";
        case "MUL":
          return "times";
        case "DIV":
          return "div";
        case "PERC":
          return "rem";
        case "PLUSPLUS":
          return "inc";
        case "MINUSMINUS":
          return "dec";
        case "EXCL":
          return "not";
        case "EQEQ":
          return "equals";
        case "GT":
        case "LT":
        case "GTEQ":
        case "LTEQ":
          return "compareTo";
        case "IN_KEYWORD":
        case "NOT_IN":
          return "contains";
        case "RANGE":
          return "rangeTo";
        case "RANGE_UNTIL":
          return "rangeUntil";
        case "PLUSEQ":
          return "plusAssign";
        case "MINUSEQ":
          return "minusAssign";
        case "MULTEQ":
          return "timesAssign";
        case "DIVEQ":
          return "divAssign";
        case "PERCEQ":
          return "remAssign";
        default:
          return null;
      }
    }

    /** Get the type of an expression using simple heuristics. */
    private String getSimpleExpressionType(KtExpression expression) {
      if (expression instanceof KtSimpleNameExpression) {
        KtSimpleNameExpression simpleExpr = (KtSimpleNameExpression) expression;
        String name = simpleExpr.getReferencedName();

        if (fqImportByNameOrAlias.containsKey(name)) {
          return fqImportByNameOrAlias.get(name).toString();
        }

        switch (name) {
          case "String":
            return "java.lang.String";
          case "Int":
            return "java.lang.Integer";
          case "Double":
            return "java.lang.Double";
          case "Boolean":
            return "java.lang.Boolean";
          case "List":
            return "java.util.List";
          case "Set":
            return "java.util.Set";
          case "Map":
            return "java.util.Map";
          default:
            return name;
        }
      }

      return expression.getText();
    }

    private FqName packageRelativeName(FqName name, KtFile file) {
      return FqNamesUtilKt.tail(name, file.getPackageFqName());
    }

    /** Returns true if this simple name is PascalCase. */
    private boolean isLikelyClassName(String name) {
      if (name.isEmpty() || !firstLetterIsUppercase(name)) {
        return false;
      }
      // If the name is all uppercase, assume it's a constant. At worst, we'll still
      // import the package, which seems a safer default than assuming it's a class
      // that we then can't find.
      for (int i = 1; i < name.length(); i++) {
        char c = name.charAt(i);
        if (Character.isLetter(c) && Character.isLowerCase(c)) {
          return true;
        }
      }
      return false;
    }

    private boolean firstLetterIsUppercase(String value) {
      for (int i = 0; i < value.length(); i++) {
        char c = value.charAt(i);
        if (Character.isLetter(c)) {
          return Character.isUpperCase(c);
        }
      }
      return false;
    }

    private Optional<KtNamedFunction> retrievePossibleMainFunction(KtObjectDeclaration object) {
      KtClassBody body = object.getBody();
      for (KtNamedFunction function : body.getFunctions()) {
        if (hasMainFunctionShape(function)) {
          return Optional.of(function);
        }
      }
      return Optional.empty();
    }

    private boolean hasMainFunctionShape(KtNamedFunction function) {
      // TODO: Check return type
      if (!"main".equals(function.getName())
          || function.hasModifier(KtTokens.INTERNAL_KEYWORD)
          || function.hasModifier(KtTokens.PROTECTED_KEYWORD)
          || function.hasModifier(KtTokens.PRIVATE_KEYWORD)) {
        return false;
      }
      List<KtParameter> parameters = function.getValueParameters();
      if (parameters.size() > 1) {
        return false;
      } else if (parameters.size() == 1) {
        KtParameter parameter = parameters.get(0);
        boolean isValidMainArgs =
            (parameter.isVarArg() && parameter.getTypeReference().getTypeText().equals("String"))
                || parameter.getTypeReference().getTypeText().equals("Array<String>");
        if (!isValidMainArgs) {
          return false;
        }
      }
      return true;
    }

    private boolean isVisible() {
      return !visibilityStack.contains(Visibility.PRIVATE);
    }

    private boolean isJvmStatic(KtAnnotated annotatedThing) {
      return annotatedThing.getAnnotationEntries().stream()
          .anyMatch(entry -> entry.getTypeReference().getTypeText().equals("JvmStatic"));
    }

    private void addExportedTypeIfNeeded(KtTypeReference theType) {
      KtTypeElement typeElement = getRootType(theType);
      Optional<String> maybeQualifiedType = tryGetFullyQualifiedName(typeElement);
      // TODO: Check for java and Kotlin standard library types.
      maybeQualifiedType.ifPresent(packageData.exportedTypes::add);
    }

    private KtTypeElement getRootType(KtTypeReference typeReference) {
      KtTypeElement typeElement = typeReference.getTypeElement();
      if (typeElement instanceof KtNullableType) {
        KtNullableType nullableType = (KtNullableType) typeElement;
        return nullableType.getInnerType();
      }
      return typeElement;
    }

    private Optional<String> tryGetFullyQualifiedName(KtTypeElement typeElement) {
      if (typeElement instanceof KtUserType) {
        KtUserType userType = (KtUserType) typeElement;
        String identifier = userType.getReferencedName();
        if (identifier.contains(".")) {
          return Optional.of(identifier);
        } else {
          if (fqImportByNameOrAlias.containsKey(identifier)) {
            return Optional.of(fqImportByNameOrAlias.get(identifier).toString());
          } else {
            return Optional.empty();
          }
        }
      } else {
        return Optional.empty();
      }
    }

    private FqName javaClassNameForKtFile(KtFile file) {
      FqName filePackage = file.getPackageFqName();
      String outerClassName = NameUtils.getScriptNameForFile(file.getName()) + "Kt";
      return filePackage.child(Name.identifier(outerClassName));
    }

    /** Get the string representation of a type reference. */
    private String getTypeString(KtTypeReference typeRef) {
      if (typeRef == null) {
        return "unknown";
      }

      // Get the text representation and try to resolve it
      String typeText = typeRef.getText();

      // Try to resolve to fully qualified name if it's in imports
      if (fqImportByNameOrAlias.containsKey(typeText)) {
        return fqImportByNameOrAlias.get(typeText).toString();
      }

      // For built-in types like String, Int, etc., add java.lang prefix if needed
      if (typeText.equals("String")) {
        return "java.lang.String";
      } else if (typeText.equals("Int")) {
        return "java.lang.Integer";
      } else if (typeText.equals("Double")) {
        return "java.lang.Double";
      } else if (typeText.equals("Boolean")) {
        return "java.lang.Boolean";
      }

      // Return as-is for now (could be a local class or unresolved type)
      return typeText;
    }

    private FqName getFunctionFqName(KtNamedFunction function) {
      String functionName = function.getName();

      // Check if it's a top-level function
      if (function.isTopLevel()) {
        // Top-level functions belong to the file's generated class
        FqName fileFqName = javaClassNameForKtFile(function.getContainingKtFile());
        return fileFqName.child(Name.identifier(functionName));
      }

      // Check if it's inside a class
      if (function.getParent().getParent() instanceof KtClass) {
        KtClass clazz = (KtClass) function.getParent().getParent();
        FqName classFqName = clazz.getFqName();
        return classFqName.child(Name.identifier(functionName));
      }

      // Check if it's inside an object
      if (function.getParent().getParent() instanceof KtObjectDeclaration) {
        KtObjectDeclaration object = (KtObjectDeclaration) function.getParent().getParent();
        FqName objectFqName = object.getFqName();
        return objectFqName.child(Name.identifier(functionName));
      }

      // Fallback: use the file package and function name
      FqName packageFqName = function.getContainingKtFile().getPackageFqName();
      return packageFqName.child(Name.identifier(functionName));
    }

    /** Get the fully qualified name for a property. */
    private FqName getPropertyFqName(KtProperty property) {
      String propertyName = property.getName();

      // Check if it's a top-level property
      if (property.isTopLevel()) {
        // Top-level properties belong to the file's generated class
        FqName fileFqName = javaClassNameForKtFile(property.getContainingKtFile());
        return fileFqName.child(Name.identifier(propertyName));
      }

      // Check if it's inside a class
      if (property.getParent().getParent() instanceof KtClass) {
        KtClass clazz = (KtClass) property.getParent().getParent();
        FqName classFqName = clazz.getFqName();
        return classFqName.child(Name.identifier(propertyName));
      }

      // Check if it's inside an object
      if (property.getParent().getParent() instanceof KtObjectDeclaration) {
        KtObjectDeclaration object = (KtObjectDeclaration) property.getParent().getParent();
        FqName objectFqName = object.getFqName();
        return objectFqName.child(Name.identifier(propertyName));
      }

      // Fallback: use the file package and property name
      FqName packageFqName = property.getContainingKtFile().getPackageFqName();
      return packageFqName.child(Name.identifier(propertyName));
    }

    /** Determine the delegate type from a delegate expression. */
    private String getDelegateType(KtExpression delegateExpression) {
      String expressionText = delegateExpression.getText();

      // Handle common delegate patterns
      if (expressionText.startsWith("lazy")) {
        return "kotlin.Lazy";
      } else if (expressionText.contains("observable")) {
        return "kotlin.properties.ObservableProperty";
      } else if (expressionText.contains("vetoable")) {
        return "kotlin.properties.VetoableProperty";
      } else if (expressionText.contains("notNull")) {
        return "kotlin.properties.NotNullVar";
      } else if (expressionText.contains("Delegates.")) {
        // Extract delegate type from Delegates.xxx() calls
        if (expressionText.contains("Delegates.observable")) {
          return "kotlin.properties.ObservableProperty";
        } else if (expressionText.contains("Delegates.vetoable")) {
          return "kotlin.properties.VetoableProperty";
        } else if (expressionText.contains("Delegates.notNull")) {
          return "kotlin.properties.NotNullVar";
        }
      }

      // Try to resolve the delegate expression type using simple heuristics
      if (delegateExpression instanceof KtCallExpression) {
        KtCallExpression callExpr = (KtCallExpression) delegateExpression;
        KtExpression calleeExpr = callExpr.getCalleeExpression();
        if (calleeExpr instanceof KtSimpleNameExpression) {
          KtSimpleNameExpression simpleExpr = (KtSimpleNameExpression) calleeExpr;
          String calleeName = simpleExpr.getReferencedName();

          // Try to resolve from imports
          if (fqImportByNameOrAlias.containsKey(calleeName)) {
            return fqImportByNameOrAlias.get(calleeName).toString();
          }

          // Return the simple name as a fallback
          return calleeName;
        }
      }

      // Fallback: return the expression text as-is
      return expressionText;
    }

    private void pushState(KtElement element) {
      if (element instanceof KtModifierListOwner) {
        KtModifierListOwner modifiedThing = (KtModifierListOwner) element;
        pushVisibility(modifiedThing);
      }
    }

    private void popState(KtElement element) {
      if (element instanceof KtModifierListOwner) {
        popVisibility();
      }
    }

    private void pushVisibility(KtModifierListOwner modifiedThing) {
      if (modifiedThing.hasModifier(KtTokens.PROTECTED_KEYWORD)) {
        visibilityStack.push(Visibility.PROTECTED);
      } else if (modifiedThing.hasModifier(KtTokens.INTERNAL_KEYWORD)) {
        visibilityStack.push(Visibility.INTERNAL);
      } else if (modifiedThing.hasModifier(KtTokens.PRIVATE_KEYWORD)) {
        visibilityStack.push(Visibility.PRIVATE);
      } else {
        visibilityStack.push(Visibility.PUBLIC);
      }
    }

    private void popVisibility() {
      visibilityStack.pop();
    }
  }

  private enum Visibility {
    PUBLIC,
    PROTECTED,
    INTERNAL,
    PRIVATE,
  }
}

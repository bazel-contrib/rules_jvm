package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.nio.file.Path;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Optional;
import java.util.Stack;
import java.util.stream.Collectors;

// import org.jetbrains.kotlin.resolve.lazy.JvmResolveUtil;
import org.jetbrains.kotlin.cli.common.messages.MessageCollector;
import org.jetbrains.kotlin.config.CommonConfigurationKeys;
import org.jetbrains.kotlin.analyzer.AnalysisResult;
import org.jetbrains.kotlin.cli.jvm.compiler.NoScopeRecordCliBindingTrace;
import org.jetbrains.kotlin.cli.jvm.compiler.TopDownAnalyzerFacadeForJVM;
import org.jetbrains.kotlin.cli.jvm.compiler.EnvironmentConfigFiles;
import org.jetbrains.kotlin.cli.jvm.compiler.KotlinCoreEnvironment;
import org.jetbrains.kotlin.config.CompilerConfiguration;
import org.jetbrains.kotlin.lexer.KtTokens;
import org.jetbrains.kotlin.name.FqName;
import org.jetbrains.kotlin.name.FqNamesUtilKt;
import org.jetbrains.kotlin.name.NameUtils;
import org.jetbrains.kotlin.name.Name;
import org.jetbrains.kotlin.psi.KotlinReferenceProvidersService;
import org.jetbrains.kotlin.psi.KtAnnotated;
import org.jetbrains.kotlin.psi.KtClass;
import org.jetbrains.kotlin.psi.KtClassBody;
import org.jetbrains.kotlin.psi.KtElement;
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
import org.jetbrains.kotlin.psi.KtSimpleNameExpression;
import org.jetbrains.kotlin.psi.KtTreeVisitorVoid;
import org.jetbrains.kotlin.psi.KtTypeElement;
import org.jetbrains.kotlin.psi.KtTypeReference;
import org.jetbrains.kotlin.psi.KtUserType;
import org.jetbrains.kotlin.resolve.BindingContext;
import org.jetbrains.kotlin.types.KotlinType;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.openapi.vfs.VirtualFileManager;
import com.intellij.psi.PsiManager;
import com.intellij.psi.PsiReference;
import com.intellij.psi.search.GlobalSearchScope;

public class KtParser {
    private static final Logger logger = LoggerFactory.getLogger(GrpcServer.class);

    private final CompilerConfiguration compilerConf = createCompilerConfiguration();
    private final KotlinCoreEnvironment env = KotlinCoreEnvironment.createForProduction(
        Disposer.newDisposable(),
        compilerConf,
        EnvironmentConfigFiles.JVM_CONFIG_FILES);
    private final VirtualFileManager vfm = VirtualFileManager.getInstance();
    private final PsiManager psiManager = PsiManager.getInstance(env.getProject());

    public ParsedPackageData parseClasses(List<Path> files) {
        KtFileVisitor visitor = new KtFileVisitor();
        List<VirtualFile> virtualFiles = files.stream().map(f -> vfm.findFileByNioPath(f)).toList();

        for (VirtualFile virtualFile : virtualFiles) {
            if (virtualFile == null) {
                throw new IllegalArgumentException("File not found: " + files.get(0));
            }
            KtFile ktFile = (KtFile) psiManager.findFile(virtualFile);

            // AnalysisResult analysis = TopDownAnalyzerFacadeForJVM.analyzeFilesWithJavaIntegration(
            //     env.getProject(),
            //     List.of(ktFile),
            //     new NoScopeRecordCliBindingTrace(),
            //     compilerConf,
            //     env::createPackagePartProvider
            //     // () -> env.createPackagePartProvider(
            //     //     GlobalSearchScope.filesScope(env.getProject(), List.of(file)))
            // );
            // // AnalysisResult analysis = JvmResolveUtil.analyze(ktFile, env);
            // BindingContext context = analysis.getBindingContext();

            logger.debug("ktFile: {}", ktFile);
            logger.debug("name: {}", ktFile.getName());
            logger.debug("script: {}", ktFile.getScript());
            logger.debug("declarations: {}", ktFile.getDeclarations());
            logger.debug("classes: {}", Arrays.stream(ktFile.getClasses()).map(Object::toString).collect(Collectors.joining(", ")));
            logger.debug("import directives: {}", ktFile.getImportDirectives());
            logger.debug("import list: {}", ktFile.getImportList());

            ktFile.accept(visitor);
        }
        
        return visitor.packageData;
    }

    private static CompilerConfiguration createCompilerConfiguration() {
        CompilerConfiguration conf = new CompilerConfiguration();
        conf.put(CommonConfigurationKeys.MODULE_NAME, "bazel-module");

        // conf.put(CommonConfigurationKeys.MESSAGE_COLLECTOR_KEY, MessageCollector.Companion.getNONE());
        return conf;
    }

    public static class KtFileVisitor extends KtTreeVisitorVoid {
        final ParsedPackageData packageData = new ParsedPackageData();

        private Stack<Visibility> visibilityStack = new Stack<>();
        private HashMap<String, FqName> fqImportByNameOrAlias = new HashMap<>();

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
                    // If it's not a wildcard import and lacks an obvious class name, assume it's a function in a package.
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

            super.visitProperty(property);
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

            if (hasMainFunctionShape(function)) {
                if (function.isTopLevel()) {
                    FqName outerClassFqName = javaClassNameForKtFile(function.getContainingKtFile());
                    String outerClassName = outerClassFqName.shortName().asString();
                    packageData.mainClasses.add(outerClassName);
                } else if (function.getParent().getParent() instanceof KtObjectDeclaration) {
                    // The parent is a class/object body, then the next parent should be a class or object.
                    KtObjectDeclaration object = (KtObjectDeclaration) function.getParent().getParent();
                    FqName relativeFqName = packageRelativeName(object.getFqName(), object.getContainingKtFile());
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
            super.visitNamedFunction(function);
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

            // PsiReference[] references = referenceService.getReferences(
            //     reference
            // );
            // for (PsiReference psiReference : references) {
            //     logger.info("\tResolved reference: " + psiReference.toString());
            // }
            // var ktType = context.get(BindingContext.TYPE, reference);
            // logger.info("\tResolved KotlinType: " + ktType.toString());

            super.visitTypeReference(reference);
        }

        @Override
        public void visitReferenceExpression(KtReferenceExpression reference) {
            logger.debug("Reference expression: " + reference.getText());

            super.visitReferenceExpression(reference);
        }

        @Override
        public void visitSimpleNameExpression(KtSimpleNameExpression expression) {
            logger.debug("Simple name expression: " + expression.getText());
            logger.debug("\tReferenced name: " + expression.getReferencedName());

            // PsiReference[] references = referenceService.getReferences(
            //     expression.getReferencedNameElement()
            // );
            // for (PsiReference psiReference : references) {
            //     logger.info("\tResolved reference for named element: " + psiReference.toString());
            // }

            // references = referenceService.getReferences(
            //     expression.getIdentifier()
            // );
            // for (PsiReference psiReference : references) {
            //     logger.info("\tResolved reference for identifier: " + psiReference.toString());
            // }

            // KotlinType ktType = context.getType(expression);
            // logger.info("\tResolved KotlinType: " + ktType.toString());

            super.visitSimpleNameExpression(expression);
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
            if (!function.getName().equals("main") || function.hasModifier(KtTokens.INTERNAL_KEYWORD) || function.hasModifier(KtTokens.PROTECTED_KEYWORD) || function.hasModifier(KtTokens.PRIVATE_KEYWORD)) {
                return false;
            }
            List<KtParameter> parameters = function.getValueParameters();
            if (parameters.size() > 1) {
                return false;
            } else if (parameters.size() == 1) {
                KtParameter parameter = parameters.get(0);
                boolean isValidMainArgs = (parameter.isVarArg() && parameter.getTypeReference().getTypeText().equals("String"))
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
            return annotatedThing.getAnnotationEntries().stream().anyMatch(
                entry -> entry.getTypeReference().getTypeText().equals("JvmStatic"));
        }

        private void addExportedTypeIfNeeded(KtTypeReference theType) {
            KtTypeElement typeElement = getRootType(theType);
            Optional<String> maybeQualifiedType = tryGetFullyQualifiedName(typeElement);
            maybeQualifiedType.ifPresent(
                qualifiedType -> {
                    // TODO: Check for java and Kotlin standard library types.
                    packageData.exportedTypes.add(qualifiedType);
                }
            );
        }

        private KtTypeElement getRootType(KtTypeReference typeReference) {
            KtTypeElement typeElement = typeReference.getTypeElement();
            if (typeElement instanceof KtNullableType nullableType) {
                return nullableType.getInnerType();
            }
            return typeElement;
        }

        private Optional<String> tryGetFullyQualifiedName(KtTypeElement typeElement) {
            if (typeElement instanceof KtUserType userType) {
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

        private void pushState(KtElement element) {
            if (element instanceof KtModifierListOwner modifiedThing) {
                pushVisibility(modifiedThing);
            }
        }

        private void popState(KtElement element) {
            if (element instanceof KtModifierListOwner modifiedThing) {
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

    private static enum Visibility {
        PUBLIC,
        PROTECTED,
        INTERNAL,
        PRIVATE,
    }
}
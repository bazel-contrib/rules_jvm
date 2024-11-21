package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.nio.file.Path;
import java.util.List;
import java.util.Optional;

import com.intellij.openapi.Disposable;
import com.intellij.openapi.util.Disposer;
import com.intellij.openapi.vfs.VirtualFileManager;
import com.intellij.openapi.vfs.VirtualFile;

import com.intellij.psi.PsiManager;
import io.grpc.Metadata;
import io.grpc.StatusRuntimeException;
import io.grpc.Status;
import org.jetbrains.kotlin.config.CompilerConfiguration;
import org.jetbrains.kotlin.cli.jvm.compiler.EnvironmentConfigFiles;
import org.jetbrains.kotlin.cli.jvm.compiler.KotlinCoreEnvironment;
import org.jetbrains.kotlin.name.FqName;
import org.jetbrains.kotlin.name.FqNamesUtilKt;
import org.jetbrains.kotlin.name.Name;
import org.jetbrains.kotlin.name.NameUtils;
import org.jetbrains.kotlin.lexer.KtTokens;
import org.jetbrains.kotlin.psi.KtAnnotated;
import org.jetbrains.kotlin.psi.KtClass;
import org.jetbrains.kotlin.psi.KtClassBody;
import org.jetbrains.kotlin.psi.KtClassOrObject;
import org.jetbrains.kotlin.psi.KtObjectDeclaration;
import org.jetbrains.kotlin.psi.KtFile;
import org.jetbrains.kotlin.psi.KtImportDirective;
import org.jetbrains.kotlin.psi.KtNamedFunction;
import org.jetbrains.kotlin.psi.KtNullableType;
import org.jetbrains.kotlin.psi.KtPackageDirective;
import org.jetbrains.kotlin.psi.KtParameter;
import org.jetbrains.kotlin.psi.KtTreeVisitorVoid;
import org.jetbrains.kotlin.psi.KtTypeElement;
import org.jetbrains.kotlin.psi.KtTypeReference;
import org.jetbrains.kotlin.psi.KtUserType;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

public class KtParser {
    private final KotlinCoreEnvironment env = KotlinCoreEnvironment.createForProduction(
        Disposer.newDisposable(),
        CompilerConfiguration.EMPTY,
        EnvironmentConfigFiles.JVM_CONFIG_FILES);
    private final VirtualFileManager vfm = VirtualFileManager.getInstance();
    private final PsiManager psiManager = PsiManager.getInstance(env.getProject());

    public ParsedPackageData parseClasses(List<Path> files) {
        VirtualFile file = vfm.findFileByNioPath(files.get(0));
        if (file == null) {
            throw new IllegalArgumentException("File not found: " + files.get(0));
        }

        KtFile ktFile = (KtFile) psiManager.findFile(file);

        System.out.println("ktFile: " + ktFile);
        System.out.println("name: " + ktFile.getName());
        System.out.println("script: " + ktFile.getScript());
        System.out.println("declarations: " + ktFile.getDeclarations());
        System.out.println("classes: " + Arrays.stream(ktFile.getClasses()).map(Object::toString).collect(Collectors.joining(", ")));
        System.out.println("import directives: " + ktFile.getImportDirectives());
        System.out.println("import list: " + ktFile.getImportList());

        KtFileVisitor visitor = new KtFileVisitor();
        ktFile.accept(visitor);
        
        return visitor.packageData;
    }

    public static class KtFileVisitor extends KtTreeVisitorVoid {
        final ParsedPackageData packageData = new ParsedPackageData();

        private HashMap<String, FqName> fqImportByNameOrAlias = new HashMap<>();

        @Override
        public void visitPackageDirective(KtPackageDirective packageDirective) {
            packageData.packages.add(packageDirective.getQualifiedName());
            super.visitPackageDirective(packageDirective);
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
                if (!isLikelyClassName(className.shortName().toString())) {
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
                    // If it's a wildcard import with no PascalCase component, assume it's a package.
                    packageData.usedPackagesWithoutSpecificTypes.add(importName.asString());
                } else {
                    // If it's not a wildcard import and lacks a PascalCase component, assume it's a function in a package.
                    packageData.usedPackagesWithoutSpecificTypes.add(importName.parent().asString());
                }
            }
            super.visitImportDirective(importDirective);
        }

        @Override
        public void visitClass(KtClass clazz) {
            if (clazz.isLocal() || clazz.hasModifier(KtTokens.PRIVATE_KEYWORD)) {
                super.visitClass(clazz);
                return;
            }
            packageData.perClassData.put(clazz.getFqName().toString(), new PerClassData());
            super.visitClass(clazz);
        }

        @Override
        public void visitObjectDeclaration(KtObjectDeclaration object) {
            if (object.isLocal() || object.hasModifier(KtTokens.PRIVATE_KEYWORD)) {
                super.visitObjectDeclaration(object);
                return;
            }

            packageData.perClassData.put(object.getFqName().toString(), new PerClassData());

            Optional<KtNamedFunction> maybeMainFunction = retrievePossibleMainFunction(object);
            maybeMainFunction.ifPresent(
                mainFunction -> {
                    FqName relativeFqName = packageRelativeName(object.getFqName(), object.getContainingKtFile());
                    if (isStatic(mainFunction)) {
                        packageData.mainClasses.add(relativeFqName.parent().toString());
                    } else {
                        packageData.mainClasses.add(relativeFqName.asString());
                    }
                }
            );
            super.visitObjectDeclaration(object);
        }

        // TODO: Check for public top-level constants.

        @Override
        public void visitNamedFunction(KtNamedFunction function) {
            if (function.isLocal() || function.hasModifier(KtTokens.PRIVATE_KEYWORD)) {
                super.visitNamedFunction(function);
                return;
            }

            if (function.isTopLevel()) {
                FqName filePackage = function.getContainingKtFile().getPackageFqName();
                String outerClassName = NameUtils.getScriptNameForFile(function.getContainingKtFile().getName()) + "Kt";
                FqName outerClassFqName = filePackage.child(Name.identifier(outerClassName));
                packageData.perClassData.put(outerClassFqName.toString(), new PerClassData());
                if (hasMainFunctionShape(function)) {
                    packageData.mainClasses.add(outerClassName);
                }
            }

            KtTypeReference returnType = function.getTypeReference();
            if (returnType != null) {
                KtTypeElement typeElement = getRootType(returnType);
                Optional<String> maybeQualifiedType = tryGetFullyQualifiedName(typeElement);
                maybeQualifiedType.ifPresent(
                    qualifiedType -> {
                        // TODO: Check for java and Kotlin standard library types.
                        packageData.exportedTypes.add(qualifiedType);
                    }
                );
            }
            super.visitNamedFunction(function);
        }

        private FqName packageRelativeName(FqName name, KtFile file) {
            return FqNamesUtilKt.tail(name, file.getPackageFqName());
        }

        /** Returns true if this simple name is PascalCase. */
        private boolean isLikelyClassName(String name) {
            if (name.isEmpty() || !Character.isUpperCase(name.charAt(0))) {
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
            // TODO: Use lexer.KtTokens to check for public modifier.
            if (!function.getName().equals("main")) {
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

        private boolean isStatic(KtAnnotated annotatedThing) {
            return annotatedThing.getAnnotationEntries().stream().anyMatch(
                entry -> entry.getTypeReference().getTypeText().equals("JvmStatic"));
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
                        // TODO: Search the import declarations for this type.
                        return Optional.empty();
                    }
                }
            } else {
                return Optional.empty();
            }
        }
    }
}
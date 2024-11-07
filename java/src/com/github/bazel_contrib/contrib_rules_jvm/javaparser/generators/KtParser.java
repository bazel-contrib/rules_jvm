package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.nio.file.Path;
import java.util.List;

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
import org.jetbrains.kotlin.name.Name;
import org.jetbrains.kotlin.psi.KtClass;
import org.jetbrains.kotlin.psi.KtClassBody;
import org.jetbrains.kotlin.psi.KtClassOrObject;
import org.jetbrains.kotlin.psi.KtObjectDeclaration;
import org.jetbrains.kotlin.psi.KtFile;
import org.jetbrains.kotlin.psi.KtImportDirective;
import org.jetbrains.kotlin.psi.KtNamedFunction;
import org.jetbrains.kotlin.psi.KtPackageDirective;
import org.jetbrains.kotlin.psi.KtParameter;
import org.jetbrains.kotlin.psi.KtTreeVisitorVoid;
import java.util.Arrays;
import java.util.List;
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
                    packageData.usedTypes.add(pathSegments.subList(0, segmentCount)
                        .stream().map(Name::toString).collect(Collectors.joining(".")));
                    break;
                }
            }
            if (!foundClass) {
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
        public void visitClassOrObject(KtClassOrObject classOrObject) {
            // TODO: Check for public visibility.
            packageData.exportedTypes.add(classOrObject.getFqName().shortName().asString());

            if (classOrObject instanceof KtObjectDeclaration && hasMainFunction((KtObjectDeclaration) classOrObject)) {
                packageData.mainClasses.add(classOrObject.getFqName().shortName().asString());
            }

            super.visitClassOrObject(classOrObject);
        }

        @Override
        public void visitNamedFunction(KtNamedFunction function) {
            if (function.isTopLevel()) {
                String fileName = function.getContainingKtFile().getName();
                String outerClassName = fileName.substring(0, fileName.length() - 3) + "Kt";
                packageData.exportedTypes.add(outerClassName);
                if (hasMainFunctionShape(function)) {
                    packageData.mainClasses.add(outerClassName);
                }
            }
            super.visitNamedFunction(function);
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

        private boolean hasMainFunction(KtObjectDeclaration object) {
            // TODO: maybe check for @JvmStatic?
            KtClassBody body = object.getBody();
            for (KtNamedFunction function : body.getFunctions()) {
                if (hasMainFunctionShape(function)) {
                    return true;
                }
            }
            return false;
        }

        private boolean hasMainFunctionShape(KtNamedFunction function) {
            // TODO: Use KtLexer to check for public modifier.
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
    }
}
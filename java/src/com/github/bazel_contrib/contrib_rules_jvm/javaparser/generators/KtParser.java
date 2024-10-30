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
import org.jetbrains.kotlin.psi.KtFile;
import org.jetbrains.kotlin.psi.KtPackageDirective;
import org.jetbrains.kotlin.psi.KtTreeVisitorVoid;
import java.util.List;

public class KtParser {
    private final KotlinCoreEnvironment env = KotlinCoreEnvironment.createForProduction(
        Disposer.newDisposable(),
        CompilerConfiguration.EMPTY,
        EnvironmentConfigFiles.JVM_CONFIG_FILES);
    private final VirtualFileManager vfm = VirtualFileManager.getInstance();
    private final PsiManager psiManager = PsiManager.getInstance(env.getProject());

    public ParsedPackageData parseClasses(Path directory, List<String> files) {
        VirtualFile file = vfm.findFileByNioPath(directory.resolve(files.get(0)));

        KtFile ktFile = (KtFile) psiManager.findFile(file);

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
    }
}
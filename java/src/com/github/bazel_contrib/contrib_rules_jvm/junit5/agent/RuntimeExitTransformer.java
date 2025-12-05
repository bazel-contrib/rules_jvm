package com.github.bazel_contrib.contrib_rules_jvm.junit5.agent;

import java.lang.instrument.ClassFileTransformer;
import java.security.ProtectionDomain;
import org.objectweb.asm.ClassReader;
import org.objectweb.asm.ClassVisitor;
import org.objectweb.asm.ClassWriter;
import org.objectweb.asm.MethodVisitor;
import org.objectweb.asm.Opcodes;

public class RuntimeExitTransformer implements ClassFileTransformer {

  private static final String RUNTIME_CLASS = "java/lang/Runtime";
  private static final String EXIT_METHOD = "exit";
  private static final String HALT_METHOD = "halt";

  // Duplicated from SystemExitAgent since this class is loaded by an isolated classloader
  private static final String PREVENT_EXIT_PROPERTY = "bazel.junit5runner.preventSystemExit";

  @Override
  public byte[] transform(
      ClassLoader loader,
      String className,
      Class<?> classBeingRedefined,
      ProtectionDomain protectionDomain,
      byte[] classfileBuffer) {

    if (!RUNTIME_CLASS.equals(className)) {
      return null;
    }

    try {
      ClassReader reader = new ClassReader(classfileBuffer);
      ClassWriter writer = new ClassWriter(reader, ClassWriter.COMPUTE_FRAMES);
      ClassVisitor visitor = new RuntimeClassVisitor(writer);
      reader.accept(visitor, 0);
      return writer.toByteArray();
    } catch (Exception e) {
      System.err.println("Failed to transform Runtime class: " + e.getMessage());
      return null;
    }
  }

  private static class RuntimeClassVisitor extends ClassVisitor {
    RuntimeClassVisitor(ClassWriter writer) {
      super(Opcodes.ASM9, writer);
    }

    @Override
    public MethodVisitor visitMethod(
        int access, String name, String descriptor, String signature, String[] exceptions) {
      MethodVisitor mv = super.visitMethod(access, name, descriptor, signature, exceptions);
      if ((EXIT_METHOD.equals(name) || HALT_METHOD.equals(name)) && "(I)V".equals(descriptor)) {
        return new ExitMethodVisitor(mv);
      }
      return mv;
    }
  }

  private static class ExitMethodVisitor extends MethodVisitor {
    ExitMethodVisitor(MethodVisitor mv) {
      super(Opcodes.ASM9, mv);
    }

    @Override
    public void visitCode() {
      super.visitCode();
      // Insert at the beginning of exit(int) / halt(int):
      //   if (System.getProperty("bazel.junit5runner.preventSystemExit") != null) {
      //     throw new SecurityException("System.exit is not allowed");
      //   }
      mv.visitLdcInsn(PREVENT_EXIT_PROPERTY);
      mv.visitMethodInsn(
          Opcodes.INVOKESTATIC,
          "java/lang/System",
          "getProperty",
          "(Ljava/lang/String;)Ljava/lang/String;",
          false);
      org.objectweb.asm.Label continueLabel = new org.objectweb.asm.Label();
      mv.visitJumpInsn(Opcodes.IFNULL, continueLabel);
      mv.visitTypeInsn(Opcodes.NEW, "java/lang/SecurityException");
      mv.visitInsn(Opcodes.DUP);
      mv.visitLdcInsn("System.exit is not allowed during test execution");
      mv.visitMethodInsn(
          Opcodes.INVOKESPECIAL,
          "java/lang/SecurityException",
          "<init>",
          "(Ljava/lang/String;)V",
          false);
      mv.visitInsn(Opcodes.ATHROW);
      mv.visitLabel(continueLabel);
    }
  }
}

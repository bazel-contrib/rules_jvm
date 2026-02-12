package workspace.com.gazelle.java.javaparser.generators;

import java.lang.reflect.Method;

public class MixedRefsAndPrivateInner {
    public void check(Method method) {
        // External class literal - should be detected
        if (method.isAnnotationPresent(Uninterruptible.class)) {
            // Private inner class - should NOT be in samePackageTypeReferences
            InvokeWithExceptionHandling handler = new InvokeWithExceptionHandling();
            handler.run();
        }
    }

    private static class InvokeWithExceptionHandling implements Runnable {
        @Override
        public void run() {}
    }
}

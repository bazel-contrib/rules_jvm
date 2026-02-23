package com.example.uninterruptibles;

import java.lang.reflect.Method;

public class UninterruptibleListener {
    private UninterruptibleListener() {}

    public static void run(Runnable r, Method method) {
        // Use the Uninterruptible annotation via .class literal (same-package reference)
        if (method.isAnnotationPresent(Uninterruptible.class)) {
            InvokeWithExceptionHandling handler = new InvokeWithExceptionHandling(r);
            handler.run();
        }
    }

    private static class InvokeWithExceptionHandling implements Runnable {
        private final Runnable delegate;

        InvokeWithExceptionHandling(Runnable delegate) {
            this.delegate = delegate;
        }

        @Override
        public void run() {
            try {
                delegate.run();
            } catch (Exception e) {
                throw new RuntimeException(e);
            }
        }
    }

    private static class UninterruptibleInterceptor {
        void intercept() {}
    }
}

package workspace.com.gazelle.java.javaparser.generators;

public class PrivateInnerClassExample {
    public void run() {
        InvokeWithExceptionHandling handler = new InvokeWithExceptionHandling();
        handler.run();
    }

    private static class InvokeWithExceptionHandling implements Runnable {
        @Override
        public void run() {}
    }

    private static class UninterruptibleInterceptor {
        void intercept() {}
    }
}

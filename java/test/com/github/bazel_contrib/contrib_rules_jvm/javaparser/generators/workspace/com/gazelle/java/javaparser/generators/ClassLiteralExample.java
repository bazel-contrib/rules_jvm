package workspace.com.gazelle.java.javaparser.generators;

import java.lang.reflect.Method;

public class ClassLiteralExample {
    public void check(Method method) {
        if (method.isAnnotationPresent(Uninterruptible.class)) {
            System.out.println("found");
        }
    }
}

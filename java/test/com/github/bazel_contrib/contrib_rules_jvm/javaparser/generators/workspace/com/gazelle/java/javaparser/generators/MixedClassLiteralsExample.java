package workspace.com.gazelle.java.javaparser.generators;

import java.util.function.BiFunction;

public class MixedClassLiteralsExample {
    public void findMethod(Class<?> serviceType) {
        // BiFunction is imported - should NOT be in samePackageTypeReferences
        findMethod(serviceType, BiFunction.class);
        // Uninterruptible is NOT imported - should be in samePackageTypeReferences
        findMethod(serviceType, Uninterruptible.class);
    }

    private void findMethod(Class<?> type, Class<?> param) {}
}

package workspace.com.gazelle.java.javaparser.generators;

import java.util.function.BiFunction;
import com.google.protobuf.Message;

public class ImportedClassLiteralExample {
    public void findMethod(Class<?> serviceType) {
        findStaticMethod(serviceType, "newApi",
            BiFunction.class,
            BiFunction.class,
            Message.class);
    }

    private void findStaticMethod(Class<?> type, String name, Class<?>... params) {}
}

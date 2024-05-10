package workspace.com.gazelle.java.javaparser.generators;

import lombok.Getter;
import lombok.Setter;
import com.example.NonNull;

public class AnnotationOnField {
    @Getter
    private final String someField;

    public void doSomething() {
        @NonNull String variable = "hello";
    }

    private static class InnerClass {
        @Setter
        private String canBeSet;
    }

    public enum InnerEnum {
        VARIANT;

        @Getter
        private final int size;
    }
}

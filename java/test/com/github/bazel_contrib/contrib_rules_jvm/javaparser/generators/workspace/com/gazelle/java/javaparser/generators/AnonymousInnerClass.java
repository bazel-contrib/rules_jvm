package workspace.com.gazelle.java.javaparser.generators;

import java.util.HashMap;

public class AnonymousInnerClass {
    public static final HashMap<String, String> map = new HashMap<>() {
        @javax.annotation.Nullable
        @Override
        public boolean containsValue(@org.jetbrains.annotations.Nullable Object value) {
            return true;
        }
    };
}

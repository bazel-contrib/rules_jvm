package workspace.com.gazelle.java.javaparser.generators;

import com.example.Registry;

public class ClassLiteral {
    public void register() {
        // The .class literal on MyHandler should be detected as a same-package type reference.
        Registry.register(MyHandler.class);
    }
}

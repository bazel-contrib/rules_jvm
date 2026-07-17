package workspace.com.gazelle.java.javaparser.generators;

import static com.example.Outer.Inner;

public class StaticImportNestedClass {
    // Inner should be resolvable as a type because the static import brings it into scope.
    // Current parser doesn't register static-imported class names in currentFileImports.
    Inner value;
}

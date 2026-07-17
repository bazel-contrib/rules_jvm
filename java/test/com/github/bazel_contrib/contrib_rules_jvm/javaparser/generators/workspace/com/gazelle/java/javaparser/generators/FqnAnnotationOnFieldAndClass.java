package workspace.com.gazelle.java.javaparser.generators;

// No import for com.example.ClassAnnotation or com.example.FieldAnnotation —
// they are used fully-qualified directly in the source.

@com.example.ClassAnnotation
public class FqnAnnotationOnFieldAndClass {
    @com.example.FieldAnnotation
    private String myField;
}

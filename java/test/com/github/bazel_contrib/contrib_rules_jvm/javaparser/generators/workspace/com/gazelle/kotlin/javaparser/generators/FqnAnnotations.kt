package workspace.com.gazelle.kotlin.javaparser.generators

// No imports — annotations are used fully-qualified.
// FQN annotations bypass the import handler entirely.

@com.example.annotations.ClassAnnotation
class FqnAnnotations {
    @com.example.annotations.FieldAnnotation
    val myField: String = "hello"

    @com.example.annotations.MethodAnnotation
    fun myMethod() {}
}

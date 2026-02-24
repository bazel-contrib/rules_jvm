package workspace.com.gazelle.java.javaparser.generators;

@SomeClassAnnotation
public class SamePackageAllPositions {
    SomeFieldType field;

    @SomeMethodAnnotation
    public <R extends SomeMethodBound> R transform(String input) throws SomeCheckedException {
        return null;
    }
}

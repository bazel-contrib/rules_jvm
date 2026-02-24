package workspace.com.gazelle.java.javaparser.generators;

public class SamePackageWithGenerics<T extends SomeBound> {
    public <R> R convert(T input) { return null; }
}

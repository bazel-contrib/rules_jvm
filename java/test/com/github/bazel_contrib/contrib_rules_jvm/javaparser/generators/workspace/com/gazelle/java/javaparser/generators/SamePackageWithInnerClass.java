package workspace.com.gazelle.java.javaparser.generators;

public class SamePackageWithInnerClass {
    public InnerHelper createHelper() { return new InnerHelper(); }
    public ExternalHelper getExternal() { return null; }

    public static class InnerHelper {}
}

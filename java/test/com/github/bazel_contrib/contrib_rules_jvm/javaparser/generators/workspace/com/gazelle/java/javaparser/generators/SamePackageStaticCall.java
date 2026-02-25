package workspace.com.gazelle.java.javaparser.generators;

public class SamePackageStaticCall {
    Object result = ExternalFactory.create(42);
    Object other = ExternalFactory.build("test", new Object());
}

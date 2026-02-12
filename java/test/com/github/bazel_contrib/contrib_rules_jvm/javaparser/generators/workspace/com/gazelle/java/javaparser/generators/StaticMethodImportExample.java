package workspace.com.gazelle.java.javaparser.generators;

import static org.junit.Assert.assertEquals;
import static com.google.common.base.Preconditions.checkNotNull;

public class StaticMethodImportExample {
    public void test() {
        assertEquals(1, 1);
        checkNotNull(new Object());
        // This should be a same-package reference since it's not imported
        SomeHelper helper = new SomeHelper();
    }
}

package workspace.com.gazelle.java.javaparser.generators;

import com.google.common.primitives.Ints;

public class App {
    public static int compare(int a, int b) {
        return Ints.compare(a, b);
    }

    public static void main(String... args) throws Exception {
        App app = new App();
        System.out.println("Success: " + app.compare(2, 1));
    }
}

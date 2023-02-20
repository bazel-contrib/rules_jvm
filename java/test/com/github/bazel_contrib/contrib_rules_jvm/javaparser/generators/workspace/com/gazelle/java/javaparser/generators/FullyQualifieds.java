package workspace.com.gazelle.java.javaparser.generators;

public class FullyQualifieds {
  public void fn() {
    new workspace.com.gazelle.java.javaparser.generators.DeleteBookRequest();
    workspace.com.gazelle.java.javaparser.utils.Printer.print();

    workspace.com.gazelle.java.javaparser.generators.DeleteBookResponse response =
        workspace.com.gazelle.java.javaparser.factories.Factory.create();

    // instance methods shouldn't get detected as classes (e.g. this.foo).
    this.foo.bar();

    // Anonymous variables in lambdas shouldn't be picked up as variable names - visitVariable sees
    // them as variables with null types.
    someList.map(x -> x.toString());

    this.BLAH = "beep";
    this.BEEP_BOOP = "baz";
  }
}

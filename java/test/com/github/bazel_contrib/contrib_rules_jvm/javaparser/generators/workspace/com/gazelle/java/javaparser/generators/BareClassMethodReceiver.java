package workspace.com.gazelle.java.javaparser.generators;

// No import — SamePackageHelper is a same-package class used as a bare method receiver.
// visitMethodInvocation only handles MemberSelectTree receivers (a.B.method()),
// not IdentifierTree receivers (B.method()). So SamePackageHelper.create()
// is invisible to the method-invocation detection path, and without an import,
// the same-package fallback in checkFullyQualifiedType is never reached.

public class BareClassMethodReceiver {
    public void doWork() {
        SamePackageHelper.create();
    }
}

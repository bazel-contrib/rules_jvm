package workspace.com.gazelle.kotlin.javaparser.generators

// No import — SamePackageHelper is in the same package but may be in a different
// Bazel target (split package). The KtParser relies entirely on imports for
// usedTypes; visitSimpleNameExpression sees "SamePackageHelper" and recognises
// it as a likely class name, but never adds it to usedTypes. Without a
// same-package fallback (like ClasspathParser's checkFullyQualifiedType),
// this dependency is invisible.

class BareClassMethodReceiver {
    fun doWork() {
        SamePackageHelper.create()
    }
}

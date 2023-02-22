package gazelle

// packagesKey is the name of a private attribute set on generated java_library
// rules. This attribute contains a list of package names (as type types.PackageName) it can be imported
// for. Note that the Java plugin currently uses package names, not classes, as its importable unit.
const packagesKey = "_java_packages"

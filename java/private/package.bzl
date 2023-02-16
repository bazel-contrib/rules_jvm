# Common package prefixes, in the order we want to check for them
_PREFIXES = (".com.", ".org.", ".net.", ".io.", ".ai.", ".co.", ".me.")

# By default bazel computes the name of test classes based on the
# standard Maven directory structure, which we may not always use,
# so try to compute the correct package name.
def get_package_name(prefixes = []):
    pkg = native.package_name().replace("/", ".")
    if len(prefixes) == 0:
        prefixes = _PREFIXES

    for prefix in prefixes:
        idx = pkg.find(prefix)
        if idx != -1:
            return pkg[idx + 1:] + "."

    return ""

# Converts a file name into what is hopefully a valid class name.
def get_class_name(package, src, prefixes = []):
    # Strip the suffix from the source
    idx = src.rindex(".")
    name = src[:idx].replace("/", ".")

    if len(prefixes) == 0:
        prefixes = _PREFIXES

    for prefix in prefixes:
        idx = name.find(prefix)
        if idx != -1:
            return name[idx + 1:]

    # Make sure that the package has a trailing period so it's
    # safe to add the class name. While `get_package_name` does
    # the right thing, the parameter passed by a user may not
    # so we shall check once we have `pkg` just to be safe.
    pkg = package if package else get_package_name(prefixes)
    if len(pkg) and not pkg.endswith("."):
        pkg = pkg + "."

    if pkg:
        return pkg + name
    return name

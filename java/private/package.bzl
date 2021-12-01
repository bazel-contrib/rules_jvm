# Common package prefixes, in the order we want to check for them
_PREFIXES = (".com.", ".org.", ".net.", ".io.")

# By default bazel computes the name of test classes based on the
# standard Maven directory structure, which we may not always use,
# so try to compute the correct package name.
def get_package_name():
    pkg = native.package_name().replace("/", ".")

    for prefix in _PREFIXES:
        idx = pkg.find(prefix)
        if idx != -1:
            return pkg[idx + 1:] + "."

    return ""

# Converts a file name into what is hopefully a valid class name.
def get_class_name(src):
    # Strip the suffix from the source
    idx = src.rindex(".")
    name = src[:idx].replace("/", ".")

    for prefix in _PREFIXES:
        idx = name.find(prefix)
        if idx != -1:
            return name[idx + 1:]

    pkg = get_package_name()
    if pkg:
        return pkg + name
    return name

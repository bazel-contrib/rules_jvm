load("//java:defs.bzl", "java_junit5_test", "java_library", test_suite = "create_jvm_test_suite")

def _define_junit5_test(name, **kwargs):
    duplicate_test_name = kwargs.pop("duplicate_test_name", None)

    test_name = "%s-%s" % (duplicate_test_name, name) if duplicate_test_name else name
    java_junit5_test(
        name = test_name,
        **kwargs
    )

    return test_name

def _define_library(name, **kwargs):
    java_library(
        name = name,
        **kwargs
    )

def java_test_suite(
        name,
        runner = "junit5",
        test_suffixes = ["Test.java"],
        package = None,
        **kwargs):
    test_suite(
        name,
        test_suffixes = test_suffixes,
        package = package,
        define_test = _define_junit5_test,
        define_library = _define_library,
        runner = runner,
        duplicate_test_name = name,
        **kwargs
    )

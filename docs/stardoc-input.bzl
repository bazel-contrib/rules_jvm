load(
    "//java:defs.bzl",
    _checkstyle_config = "checkstyle_config",
    _checkstyle_test = "checkstyle_test",
    _java_binary = "java_binary",
    _java_export = "java_export",
    _java_junit5_test = "java_junit5_test",
    _java_library = "java_library",
    _java_test = "java_test",
    _java_test_suite = "java_test_suite",
)

checkstyle_config = _checkstyle_config
checkstyle_test = _checkstyle_test
java_binary = _java_binary
java_export = _java_export
java_library = _java_library
java_junit5_test = _java_junit5_test
java_test = _java_test
java_test_suite = _java_test_suite

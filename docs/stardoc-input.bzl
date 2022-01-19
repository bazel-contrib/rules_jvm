load(
    "//java:defs.bzl",
    _checkstyle_binary = "checkstyle_binary",
    _checkstyle_config = "checkstyle_config",
    _checkstyle_test = "checkstyle_test",
    _java_binary = "java_binary",
    _java_export = "java_export",
    _java_junit5_test = "java_junit5_test",
    _java_library = "java_library",
    _java_test = "java_test",
    _java_test_suite = "java_test_suite",
    _pmd_ruleset = "pmd_ruleset",
    _pmd_test = "pmd_test",
    _spotbugs_binary = "spotbugs_binary",
    _spotbugs_config = "spotbugs_config",
    _spotbugs_test = "spotbugs_test",
)

checkstyle_binary = _checkstyle_binary
checkstyle_config = _checkstyle_config
checkstyle_test = _checkstyle_test
java_binary = _java_binary
java_export = _java_export
java_library = _java_library
java_junit5_test = _java_junit5_test
java_test = _java_test
java_test_suite = _java_test_suite
pmd_ruleset = _pmd_ruleset
pmd_test = _pmd_test
spotbugs_binary = _spotbugs_binary
spotbugs_config = _spotbugs_config
spotbugs_test = _spotbugs_test

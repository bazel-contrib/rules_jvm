load("//java/private:checkstyle.bzl", _checkstyle_test = "checkstyle_test")
load(
    "//java/private:checkstyle_config.bzl",
    _checkstyle_binary = "checkstyle_binary",
    _checkstyle_config = "checkstyle_config",
)
load("//java/private:create_jvm_test_suite.bzl", _create_jvm_test_suite = "create_jvm_test_suite")
load("//java/private:java_test_suite.bzl", _java_test_suite = "java_test_suite")
load(
    "//java/private:junit5.bzl",
    _JUNIT5_DEPS = "JUNIT5_DEPS",
    _JUNIT5_RUNTIME_DEPS = "JUNIT5_RUNTIME_DEPS",
    _JUNIT5_VINTAGE_DEPS = "JUNIT5_VINTAGE_DEPS",
    _java_junit5_test = "java_junit5_test",
    _junit5_deps = "junit5_deps",
    _junit5_jvm_flags = "junit5_jvm_flags",
    _junit5_vintage_deps = "junit5_vintage_deps",
)
load(
    "//java/private:library.bzl",
    _java_binary = "java_binary",
    _java_export = "java_export",
    _java_library = "java_library",
    _java_test = "java_test",
)
load("//java/private:pmd.bzl", _pmd_test = "pmd_test")
load("//java/private:pmd_ruleset.bzl", _pmd_binary = "pmd_binary", _pmd_ruleset = "pmd_ruleset")
load("//java/private:spotbugs.bzl", _spotbugs_test = "spotbugs_test")
load(
    "//java/private:spotbugs_config.bzl",
    _spotbugs_binary = "spotbugs_binary",
    _spotbugs_config = "spotbugs_config",
)

checkstyle_binary = _checkstyle_binary
checkstyle_config = _checkstyle_config
checkstyle_test = _checkstyle_test
create_jvm_test_suite = _create_jvm_test_suite
java_binary = _java_binary
java_export = _java_export
java_library = _java_library
java_junit5_test = _java_junit5_test
java_test = _java_test
java_test_suite = _java_test_suite
junit5_deps = _junit5_deps
junit5_jvm_flags = _junit5_jvm_flags
junit5_vintage_deps = _junit5_vintage_deps
JUNIT5_DEPS = _JUNIT5_DEPS
JUNIT5_RUNTIME_DEPS = _JUNIT5_RUNTIME_DEPS
JUNIT5_VINTAGE_DEPS = _JUNIT5_VINTAGE_DEPS
pmd_binary = _pmd_binary
pmd_ruleset = _pmd_ruleset
pmd_test = _pmd_test
spotbugs_binary = _spotbugs_binary
spotbugs_config = _spotbugs_config
spotbugs_test = _spotbugs_test

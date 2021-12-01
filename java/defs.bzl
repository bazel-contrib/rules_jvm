load("//java/private:java_test_suite.bzl", _java_test_suite = "java_test_suite")
load(
    "//java/private:junit5.bzl",
    _JUNIT5_DEPS = "JUNIT5_DEPS",
    _JUNIT5_VINTAGE_DEPS = "JUNIT5_VINTAGE_DEPS",
    _java_junit5_test = "java_junit5_test",
)

java_junit5_test = _java_junit5_test
java_test_suite = _java_test_suite
JUNIT5_DEPS = _JUNIT5_DEPS
JUNIT5_VINTAGE_DEPS = _JUNIT5_VINTAGE_DEPS

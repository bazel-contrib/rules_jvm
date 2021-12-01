load(
    "//java/private:junit5.bzl",
    _JUNIT5_DEPS = "JUNIT5_DEPS",
    _JUNIT5_VINTAGE_DEPS = "JUNIT5_VINTAGE_DEPS",
    _java_junit5_test = "java_junit5_test",
)

java_junit5_test = _java_junit5_test
JUNIT5_DEPS = _JUNIT5_DEPS
JUNIT5_VINTAGE_DEPS = _JUNIT5_VINTAGE_DEPS

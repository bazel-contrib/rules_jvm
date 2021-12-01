load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("//java/private:zip_repository.bzl", "zip_repository")

def rules_jvm_contrib_deps():
    RULES_JVM_EXTERNAL_TAG = "4.2"
    RULES_JVM_EXTERNAL_SHA = "cd1a77b7b02e8e008439ca76fd34f5b07aecb8c752961f9640dea15e9e5ba1ca"

    maybe(
        http_archive,
        name = "rules_jvm_external",
        sha256 = RULES_JVM_EXTERNAL_SHA,
        strip_prefix = "rules_jvm_external-%s" % RULES_JVM_EXTERNAL_TAG,
        url = "https://github.com/bazelbuild/rules_jvm_external/archive/%s.zip" % RULES_JVM_EXTERNAL_TAG,
        patches = [
            "@rules_jvm_contrib//java/private:make-docs-visible.patch",
        ],
        patch_args = ["-p1"],
    )

    maybe(
        zip_repository,
        name = "rules_jvm_contrib_deps",
        path = "@rules_jvm_contrib//java/private:rules_jvm_contrib_deps.zip",
    )

    maybe(
        http_archive,
        name = "apple_rules_lint",
        sha256 = "8feab4b08a958b10cb2abb7f516652cd770b582b36af6477884b3bba1f2f0726",
        strip_prefix = "apple_rules_lint-0.1.1",
        url = "https://github.com/apple/apple_rules_lint/archive/0.1.1.zip",
    )

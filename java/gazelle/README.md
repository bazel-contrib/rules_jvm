# Gazelle Java extension

This provides a  [Gazelle][] extension to generate build files for Java projects.

## Usage
In the `WORKSPACE` file set up the rules_jvm correctly:
```starlark
load("@contrib_rules_jvm//:repositories.bzl", "contrib_rules_jvm_deps", "contrib_rules_jvm_gazelle_deps")

contrib_rules_jvm_deps()

contrib_rules_jvm_gazelle_deps()

load("@contrib_rules_jvm//:setup.bzl", "contrib_rules_jvm_setup")

contrib_rules_jvm_setup()

load("@contrib_rules_jvm//:gazelle_setup.bzl", "contrib_rules_jvm_gazelle_setup")

contrib_rules_jvm_gazelle_setup()
```

In the top level `BUILD.bazel` file, setup Gazelle to use gazelle-languages binary:

```starlark
load("@bazel_gazelle//:def.bzl", "DEFAULT_LANGUAGES", "gazelle", "gazelle_binary")

# gazelle:prefix github.com/your/project
gazelle(
    name = "gazelle",
    gazelle = ":gazelle_bin",
)

gazelle_binary(
    name = "gazelle_bin",
    languages = DEFAULT_LANGUAGES + [
        "@contrib_rules_jvm//java/gazelle",
    ],
)
```

Make sure you have everything setup properly by building the gazelle binary:
`bazel build //:gazelle_bin`

To generate BUILD files:

```bash
# Run Gazelle with the java extension
bazel run //:gazelle
```

## Requirements

This gazelle plugin requires Go 1.18 or above.

## Configuration options

This Gazelle extension supports some configuration options, which are enabled by
adding comments to your root `BUILD.bazel` file. For example, to set
`java_maven_install_file`, you would add the following to your root
`BUILD.bazel` file:

```starlark
# gazelle:java_maven_install_file project/main_maven_install.json
```

See [javaconfig/config.go](javaconfig/config.go) for a list of configuration
options and their documentation.

Additionally, some configuration can only be done by flag. See the
`RegisterFlags` function in [configure.go](configure.go) for a list of these
options.

## Source code restrictions and limitations

Currently, the gazelle plugin makes the following assumptions about the code it's generating BUILD files for:
1. All code lives in a non-empty package. Source files must have a `package` declaration, and classes depended on all themselves have a `package` declaration.
1. Packages only exist in one place. Two different directories or dependencies may not contain classes which belong in the same package. The exception to this is that for each package, there may be a single test directory which uses the same package as that package's non-test directory.
1. There are no circular dependencies that extend beyond a single package. If these are present, and can't easily be removed, you may want to set `# gazelle:java_module_granularity module` in the BUILD file containing the parent-most class in the dependency cycle, which may fix the problem, but will slow down your builds. Ideally, remove dependency cycles.
1. Non-test code doesn't depend on test code.
1. Non-test code used by one package of tests either lives in the same directory as those tests, or lives in a non-test-code directory. We also detect non-test code used from another test package, if that other package doesn't have a corresponding non-test code directory, but require you to manually set the visibility on the depended-on target, because this is an unexpected set-up.
1. Package names and class/interface names follow standard java conventions; that is: package names are all lower-case, and class and interface names start with Upper Case letters.
1. Code doesn't use types which it doesn't name _only_ through unnamed means, across multiple calls. For example, if some code calls `x.foo().bar()` where the return type of `foo` is defined in another target, and the calling code explicitly uses a type from that target somewhere else. In the case of `x.foo()`, we add exports so that the caller will have access to the return type of `foo()`, but do not track dependencies on the return types across _multiple_ calls.

   This limitation could be lifted, but would require us to export all _transitively_ used symbols from every function. This would serve to add direct dependencies between lots of targets, which can slow down compilation and reduce cache hits.

   In our experience, this kind of code is rare in Java - most code tends to either introduce intermediate variables (at which point the type gets used and we detect that a dependency needs to be added), or tends to already use the targets containing the intermediate types somewhere else (at which point the dependency will already exist), but we're open to discussion about this heuristic if it poses problems for a real-world codebase.

If these assumptions are violated, the rest of the generation should still function properly, but the specific files which violate the assumptions (or depend on files which violate the assumptions) will not get complete results. We strive to emit warnings when this happens.

We are also aware of the following limitations. This list is not exhaustive, and is not intentional (i.e. if we can fix these limitations, we would like to):
1. Runtime dependencies are not detected (e.g. loading classes by reflection).

## Flags

The Java plugin for Gazelle adds the following flags to the command line options for Gazelle:

| **Name**                                      | **Default value**                                          |
|-----------------------------------------------|------------------------------------------------------------|
| java-annotation-to-attribute                  | none                                                       |
| Mapping of annotations (on test classes) to attributes which should be set for that test rule  Examples: com.example.annotations.FlakyTest=flaky=True com.example.annotations.SlowTest=timeout=\"long\"") |
| java-annotation-to-wrapper                    | none                                                       |
| Mapping of annotations (on test classes) to wrapper rules which should be used around the test rule.  
  Example: com.example.annotations.RequiresNetwork=@some//wrapper:file.bzl=requires_network")                |
| java-maven-install-file                       | "maven_install.json"                                       |
| Path of the maven_install.json file.                                                                       |


## Directives

Gazelle can be configured with directives, which are written as top-level comments in build files. Most options that can be set on the command line can also be set using directives. Some options can only be set with directives.

Directives apply in the directory where they are set and in subdirectories. This means, for example, if you set # gazelle:prefix in the build file in your project's root directory, it affects your whole project. If you set it in a subdirectory, it only affects rules in that subtree.

The following directives specific to the Java extension are recognized:

| **Directive**                                     | **Default value**                        |
|---------------------------------------------------|------------------------------------------|
| java_exclude_artifact                             | none                                     |
| Tells the resolver to disregard a given maven artifact. Used to resolve duplicate artifacts  |
| java_extension                                    | enabled                                  |
| Controls if this Java extension is enabled or not. Sub-packages inherit this value. Can be either "enabled" or "disabled". Defaults to "enabled".                                |
| java_maven_install_file                           | "maven_install.json"                     |
| Controls where the maven_install.json file is located, and named.                            |
| java_module_granularity                           | "package"                                |
| Controls whether this Java module has a module granularity or a package granularity Package granularity builds a `java_library` or `java_test_suite` for eash directory (bazel). Module graularity builds a `java_library` or `java_test_suite` for a directory and all subdirectories. This can be useful for resolving dependency loops in closely releated code. Can be either "package" or "module", defaults to "package". |
| java_test_file_suffixes                           | none                                     |
| Indicates within a test directory which files are test classes vs utility classes, based on their basename. It should be set up to match the value used for `java_test_suite`'s `test_suffixes` attribute. Accepted values are a comma-delimited list of strings.            |
| java_test_mode                                    | "suite"                                  |
| Within a test directory determines the syle of test generation. Suite generates a single `java_test_suite` for the whole directory. File generates one `java_test` rule for each test file in the directory and a `java_library` for the utility classes. Can be either "suite" or "file", defaultes to "suite". |
| java_generate_proto                               | True                                     |
| Tells the code generator to generate `java_proto_library` (and `java_library`) rules when a `proto_library` rule is present. Defaults to True. |
| java_maven_repository_name                        | "maven"                                  |
| Tells the code generator what the repository name that contains all maven dependencies is. Defaults to "maven" |
| java_annotation_processor_plugin                  | none                                     |
| Tells the code generator about specific java_plugin targets needed to process specific annotations. |
| java_resolve_to_java_exports                      | True                                     |
| Tells the code generator to favour resolving dependencies to java_exports where possible. If enabled, generated libraries will try to depend on java_exports targets that export a given package, instead of the underlying library. This allows monorepos to closely match a traditional Gradle/Maven model where subprojects are published in jars. Can be either "true" or "false". Defaults to "true". can only be set at the root of the repository. |
| java_sourceset_root                               | none                                     |
| Sourceset root explicitly marks a directory as the root of a sourceset. This provides a clear override to the auto-detection algorithm. Example: `# gazelle:java_sourceset_root my/custom/src` |
| java_strip_resources_prefix                       | none                                     |
| Strip resources prefix overrides the path-stripping behavior for resources. This is a direct way to specify the resource_strip_prefix for all resources in a directory. Example: `# gazelle:java_strip_resources_prefix my/data/config` |
| java_generate_binary                              | True                                     |
| Controls if the generator adds `java_binary` targets to the build file. If set False, no `java_binary` targets are generated for the directories, defaults to True. |
| java_search_exclude                               | none                                     |
| Excludes a directory from source root discovery when using `java_maven_layout`. This is useful for excluding build outputs, vendored code, or test fixtures. Can be repeated to exclude multiple directories. Example: `# gazelle:java_search_exclude build/generated` |

## Troubleshooting

If one forgets to run `bazel fetch @maven//...`, the code will complain and tell
you to run this command.

If one forgets to "Update the Maven mapping", they use out of date data for the
rules resolution, and the hash check will fail. An error is printed and the
resolution does not happen.

## Contibutors documentation

The following are the targets of interest:

- `//java/gazelle` implements a Gazelle extension
- `//java/gazelle/private/javaparser/cmd/javaparser-wrapper` wraps the java
  parser with an activity tracker (to stop the parser) and an adapter to prevent
  self imports.
- `//java/src/com/github/bazel_contrib/contrib_rules_jvm/javaparser/generators:Main`
  is the java parser side process

The maven integration relies on using `rules_jvm_external` at least as new as
https://github.com/bazelbuild/rules_jvm_external/pull/716

[gazelle]: https://github.com/bazelbuild/bazel-gazelle

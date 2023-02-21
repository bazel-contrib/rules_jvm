# Gazelle Java extension

This provides an experimental [Gazelle][] extension to generate build files for
Java projects.

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
1. Non-test code used by one package of tests either lives in the same directory as those tests, or lives in a non-test-code directory.

If these assumptions are violated, the rest of the generation should still function properly, but the specific files which violate the assumptions (or depend on files which violate the assumptions) will not get complete results. We strive to emit warnings when this happens.

We are also aware of the following limitations. This list is not exhaustive, and is not intentional (i.e. if we can fix these limitations, we would like to):
1. Runtime dependencies are not detected (e.g. loading classes by reflection).

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

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

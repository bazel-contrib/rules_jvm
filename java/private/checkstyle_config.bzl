load("@rules_java//java:java_binary.bzl", "java_binary")
load("@rules_java//java/common:java_info.bzl", "JavaInfo")

def checkstyle_binary(
        name,
        main_class = "com.puppycrawl.tools.checkstyle.Main",
        deps = None,
        runtime_deps = None,
        srcs = None,
        visibility = ["//visibility:public"],
        **kwargs):
    """Macro for quickly generating a `java_binary` target for use with `checkstyle_config`.

    By default, this will set the `main_class` to point to the default one used by checkstyle
    but it's ultimately a drop-replacement for straight `java_binary` target.

    At least one of `runtime_deps`, `deps`, and `srcs` must be specified so that the
    `java_binary` target will be valid.

    An example would be:

    ```starlark
    checkstyle_binary(
        name = "checkstyle_cli",
        runtime_deps = [
            artifact("com.puppycrawl.tools:checkstyle"),
        ]
    )
    ```

    Args:
      name: The name of the target
      main_class: The main class to use for checkstyle.
      deps: The deps required for compiling this binary. May be omitted.
      runtime_deps: The deps required by checkstyle at runtime. May be omitted.
      srcs: If you're compiling your own `checkstyle` binary, the sources to use.
    """
    java_binary(
        name = name,
        main_class = main_class,
        srcs = srcs,
        deps = deps,
        runtime_deps = runtime_deps,
        visibility = visibility,
        **kwargs
    )

CheckStyleInfo = provider(
    fields = {
        "checkstyle": "The checkstyle binary to use.",
        "config_file": "The config file to use.",
        "output_format": "Output Format can be plain or xml.",
    },
)

def _checkstyle_config_impl(ctx):
    return [
        DefaultInfo(
            runfiles = ctx.runfiles(ctx.files.data + [ctx.file.config_file]).merge(
                ctx.attr.checkstyle_binary[DefaultInfo].default_runfiles,
            ),
        ),
        CheckStyleInfo(
            checkstyle = ctx.executable.checkstyle_binary,
            config_file = ctx.file.config_file,
            output_format = ctx.attr.output_format,
        ),
    ]

checkstyle_config = rule(
    _checkstyle_config_impl,
    attrs = {
        "config_file": attr.label(
            doc = "The config file to use for all checkstyle tests",
            allow_single_file = True,
            mandatory = True,
        ),
        "output_format": attr.string(
            doc = "Output format to use. Defaults to plain",
            values = [
                "plain",
                "xml",
            ],
            default = "plain",
        ),
        "data": attr.label_list(
            doc = "Additional files to make available to Checkstyle such as any included XML files",
            allow_files = True,
        ),
        "checkstyle_binary": attr.label(
            doc = "Checkstyle binary to use.",
            default = "@contrib_rules_jvm//java:checkstyle_cli",
            executable = True,
            cfg = "exec",
            providers = [
                JavaInfo,
            ],
        ),
    },
    provides = [
        CheckStyleInfo,
    ],
    doc = """ Rule allowing checkstyle to be configured. This is typically
     used with the linting rules from `@apple_rules_lint` to configure how
     checkstyle should run. """,
)

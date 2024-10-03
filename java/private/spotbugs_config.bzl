def spotbugs_binary(
        name,
        main_class = "edu.umd.cs.findbugs.LaunchAppropriateUI",
        deps = None,
        runtime_deps = None,
        srcs = None,
        visibility = ["//visibility:public"],
        **kwargs):
    """Macro for quickly generating a `java_binary` target for use with `spotbugs_config`.

    By default, this will set the `main_class` to point to the default one used by spotbugs
    but it's ultimately a drop-replacement for a regular `java_binary` target.

    At least one of `runtime_deps`, `deps`, and `srcs` must be specified so that the
    `java_binary` target will be valid.

    An example would be:

    ```starlark
    spotbugs_binary(
        name = "spotbugs_cli",
        runtime_deps = [
            artifact("com.github.spotbugs:spotbugs"),
            artifact("org.slf4j:slf4j-jdk14"),
        ],
    )
    ```

    Args:
      name: The name of the target
      main_class: The main class to use for spotbugs.
      deps: The deps required for compiling this binary. May be omitted.
      runtime_deps: The deps required by spotbugs at runtime. May be omitted.
      srcs: If you're compiling your own `spotbugs` binary, the sources to use.
    """

    native.java_binary(
        name = name,
        main_class = main_class,
        srcs = srcs,
        deps = deps,
        runtime_deps = runtime_deps,
        visibility = visibility,
        **kwargs
    )

SpotBugsInfo = provider(
    fields = {
        "effort": "Effort can be min, less, default, more or max.",
        "max_rank": "Only report issues with a bug rank at least as scary as that provided.",
        "exclude_filter": "Optional filter file to use.",
        "omit_visitors": "Omit named visitors",
        "fail_on_warning": "Whether to fail on warning, or just create a report.",
        "plugin_list": "Optional list of JARs to load as plugins.",
        "binary": "The spotbugs binary to use.",
    },
)

def _spotbugs_config_impl(ctx):
    return [
        DefaultInfo(
            runfiles = ctx.attr.spotbugs_binary[DefaultInfo].default_runfiles,
        ),
        SpotBugsInfo(
            effort = ctx.attr.effort,
            max_rank = ctx.attr.max_rank,
            exclude_filter = ctx.file.exclude_filter,
            omit_visitors = ctx.attr.omit_visitors,
            plugin_list = ctx.files.plugin_list,
            fail_on_warning = ctx.attr.fail_on_warning,
            binary = ctx.executable.spotbugs_binary,
        ),
    ]

spotbugs_config = rule(
    _spotbugs_config_impl,
    doc = "Configuration used for spotbugs, typically by the `//lint` rules.",
    attrs = {
        "effort": attr.string(
            doc = "Effort can be min, less, default, more or max. Defaults to default",
            values = ["min", "less", "default", "more", "max"],
            default = "default",
        ),
        "max_rank": attr.string(
            doc = "Only report issues with a bug rank at least as scary as that provided.",
        ),
        "exclude_filter": attr.label(
            doc = "Report all bug instances except those matching the filter specified by this filter file",
            allow_single_file = True,
        ),
        "omit_visitors": attr.string_list(
            doc = "Omit named visitors.",
            default = [],
        ),
        "plugin_list": attr.label_list(
            doc = "Specify a list of plugin Jar files to load",
            allow_files = True,
        ),
        "fail_on_warning": attr.bool(
            doc = "Whether to fail on warning, or just create a report. Defaults to True",
            default = True,
        ),
        "spotbugs_binary": attr.label(
            doc = "The spotbugs binary to run.",
            default = "@contrib_rules_jvm//java:spotbugs_cli",
            executable = True,
            cfg = "exec",
            providers = [
                JavaInfo,
            ],
        ),
    },
    provides = [
        SpotBugsInfo,
    ],
)

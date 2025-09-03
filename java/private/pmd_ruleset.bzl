load("@rules_java//java:java_binary.bzl", "java_binary")

def pmd_binary(
        name,
        main_class = "net.sourceforge.pmd.cli.PmdCli",
        deps = None,
        runtime_deps = None,
        srcs = None,
        visibility = ["//visibility:public"],
        **kwargs):
    """Macro for quickly generating a `java_binary` target for use with `pmd_ruleset`.

    By default, this will set the `main_class` to point to the default one used by PMD
    but it's ultimately a drop-replacement for a regular `java_binary` target.

    At least one of `runtime_deps`, `deps`, and `srcs` must be specified so that the
    `java_binary` target will be valid.

    An example would be:

    ```starlark
    pmd_binary(
        name = "pmd",
        runtime_deps = [
            artifact("net.sourceforge.pmd:pmd-dist"),
        ],
    )
    ```

    Args:
      name: The name of the target
      main_class: The main class to use for PMD.
      deps: The deps required for compiling this binary. May be omitted.
      runtime_deps: The deps required by PMD at runtime. May be omitted.
      srcs: If you're compiling your own PMD binary, the sources to use.
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

PmdInfo = provider(
    fields = {
        "format": "The format to generate reports in",
        "rulesets": "Depset of files containing rulesets",
        "shallow": "Whether to use target outputs as part of processing",
        "binary": "The PMD binary to use",
    },
)

def _pmd_ruleset_impl(ctx):
    return [
        DefaultInfo(
            files = depset(ctx.files.rulesets),
            runfiles = ctx.attr.pmd_binary[DefaultInfo].default_runfiles,
        ),
        PmdInfo(
            rulesets = depset(ctx.files.rulesets),
            format = ctx.attr.format,
            shallow = ctx.attr.shallow,
            binary = ctx.executable.pmd_binary,
        ),
    ]

pmd_ruleset = rule(
    _pmd_ruleset_impl,
    doc = "Select a rule set for PMD tests.",
    attrs = {
        "format": attr.string(
            doc = "Generate report in the given format. One of html, text, or xml (default is xml)",
            default = "xml",
            # Per https://pmd.github.io/pmd/pmd_userdocs_report_formats.html
            values = ["codeclimate", "emacs", "html", "sarif", "csv", "ideaj", "json", "summaryhtml", "text", "textcolor", "textpad", "vbhtml", "xml", "xslt", "yahtml"],
        ),
        "rulesets": attr.label_list(
            doc = "Use these rulesets.",
            allow_files = True,
        ),
        "shallow": attr.bool(
            doc = "Use the targetted output to increase PMD's depth of processing",
            default = True,
        ),
        "pmd_binary": attr.label(
            doc = "PMD binary to use.",
            default = "//java:pmd",
            executable = True,
            cfg = "exec",
        ),
    },
    provides = [
        PmdInfo,
    ],
)

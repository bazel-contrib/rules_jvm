load(":checkstyle_config.bzl", "CheckStyleInfo")

"""
Checkstyle rule implementation
"""

def _checkstye_impl(ctx):
    # Preferred options: 1/ config on rule or 2/ config via attributes
    if ctx.attr.config != None:
        info = ctx.attr.config[CheckStyleInfo]
        config = info.config_file
        output_format = info.output_format
    elif ctx.attr.configuration_file != None:
        print("Please define a configuration using `checkstyle_config`")
        config = ctx.attr.configuration_file
        output_format = ctx.attr.output_format
    else:
        fail("Please define a configuration using `checkstyle_config` and use `@apple_rules_lint`")

    script = "\n".join([
        "#!/usr/bin/env bash",
        "RESULTS=`{lib} -f {output_format} -c {config} {srcs}|sed s:$PWD::g`".format(
            lib = ctx.executable._checkstyle_lib.short_path,
            output_format = output_format,
            config = config.short_path,
            srcs = " ".join([f.short_path for f in ctx.files.srcs]),
        ),
        "if echo \"$RESULTS\" | grep -q \"ERROR\"; then",
        "   echo \"$RESULTS\"",
        "   exit 1",
        "fi",
    ])
    out = ctx.actions.declare_file(ctx.label.name + "exec")

    ctx.actions.write(
        output = out,
        content = ctx.expand_location(script),
    )

    runfiles = ctx.runfiles(
        files = ctx.files.srcs + [ctx.executable._checkstyle_lib, config],
    )

    return [
        DefaultInfo(
            executable = out,
            runfiles = runfiles.merge(ctx.attr._checkstyle_lib[DefaultInfo].default_runfiles),
        ),
    ]

checkstyle_test = rule(
    _checkstye_impl,
    attrs = {
        "srcs": attr.label_list(
            mandatory = True,
            allow_files = True,
        ),
        "config": attr.label(
            providers = [
                [CheckStyleInfo],
            ],
        ),
        "output_format": attr.string(
            doc = "Output Format can be plain or xml. Defaults to plain",
            values = ["plain", "xml"],
            default = "plain",
        ),
        "configuration_file": attr.label(
            doc = "Configuration file. If not specified a default file is used",
            allow_single_file = True,
        ),
        "_checkstyle_lib": attr.label(
            cfg = "host",
            executable = True,
            default = "//java:checkstyle_cli",
        ),
    },
    executable = True,
    test = True,
    doc = """Use checkstyle to lint the `srcs`.""",
)

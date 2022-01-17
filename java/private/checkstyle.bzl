load(":checkstyle_config.bzl", "CheckStyleInfo")

"""
Checkstyle rule implementation
"""

def _checkstyle_impl(ctx):
    info = ctx.attr.config[CheckStyleInfo]
    config = info.config_file
    output_format = info.output_format

    script = "\n".join([
        "#!/usr/bin/env bash",
        "set -o pipefail",
        "set -e",
        "OLDPWD=$PWD",
        "cd {config_dir}".format(config_dir = config.dirname),
        "$OLDPWD/{lib} -f {output_format} -c {config} {srcs} |sed s:$OLDPWD/::g".format(
            lib = info.checkstyle.short_path,
            output_format = output_format,
            config = config.basename,
            srcs = " ".join(["$OLDPWD/" + f.short_path for f in ctx.files.srcs]),
        ),
    ])
    out = ctx.actions.declare_file(ctx.label.name + "exec")

    ctx.actions.write(
        output = out,
        content = ctx.expand_location(script),
    )

    runfiles = ctx.runfiles(
        files = ctx.files.srcs + [info.checkstyle],
    )

    return [
        DefaultInfo(
            executable = out,
            runfiles = runfiles.merge(
                ctx.attr.config[DefaultInfo].default_runfiles,
            ),
        ),
    ]

checkstyle_test = rule(
    _checkstyle_impl,
    attrs = {
        "srcs": attr.label_list(
            mandatory = True,
            allow_files = True,
        ),
        "config": attr.label(
            default = "@contrib_rules_jvm//java:checkstyle-default-config",
            providers = [
                [CheckStyleInfo],
            ],
        ),
        "output_format": attr.string(
            doc = "Output Format can be plain or xml. Defaults to plain",
            values = ["plain", "xml"],
            default = "plain",
        ),
    },
    executable = True,
    test = True,
    doc = """Use checkstyle to lint the `srcs`.""",
)

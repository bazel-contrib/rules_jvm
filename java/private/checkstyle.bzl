load("@apple_rules_lint//lint:defs.bzl", "LinterInfo")
load("@bazel_skylib//lib:paths.bzl", "paths")
load(":checkstyle_config.bzl", "CheckStyleInfo")

"""
Checkstyle rule implementation
"""

def _checkstyle_impl(ctx):
    info = ctx.attr.config[CheckStyleInfo]
    config = info.config_file
    output_format = info.output_format

    config_dir = paths.dirname(config.short_path)
    maybe_cd_config_dir = ["cd {}".format(config_dir)] if config_dir else []

    script = "\n".join([
        "#!/usr/bin/env bash",
        "set -o pipefail",
        "set +e",
        "OLDPWD=$PWD",
    ] + maybe_cd_config_dir + [
        "$OLDPWD/{lib} -o checkstyle.xml -f xml -c {config} {srcs}".format(
            lib = info.checkstyle.short_path,
            output_format = output_format,
            config = config.basename,
            srcs = " ".join(["$OLDPWD/" + f.short_path for f in ctx.files.srcs]),
        ),
        "checkstyle_exit_code=$?",
        # Apply sed to the file in place
        "sed s:$OLDPWD/::g checkstyle.xml > checkstyle-stripped.xml",
        # Run the Java XSLT transformation tool
        "$OLDPWD/{xslt_transformer} checkstyle-stripped.xml $OLDPWD/{xslt} > $XML_OUTPUT_FILE".format(
            xslt_transformer = ctx.executable._xslt_transformer.short_path,
            xslt = ctx.file.xslt.short_path,
        ),
        "exit $checkstyle_exit_code",
    ])
    out = ctx.actions.declare_file(ctx.label.name + "exec")

    ctx.actions.write(
        output = out,
        content = ctx.expand_location(script),
    )

    runfiles = ctx.runfiles(
        files = ctx.files.srcs + [info.checkstyle, ctx.file.xslt],
    ).merge(ctx.attr._xslt_transformer[DefaultInfo].default_runfiles)

    return [
        DefaultInfo(
            executable = out,
            runfiles = runfiles.merge(
                ctx.attr.config[DefaultInfo].default_runfiles,
            ),
        ),
        LinterInfo(
            language = "java",
            name = "checkstyle",
        ),
    ]

_checkstyle_test = rule(
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
        "xslt": attr.label(
            doc = "Path to the checkstyle2junit.xslt file",
            allow_single_file = True,
            default = "@contrib_rules_jvm//java:checkstyle2junit.xslt",
        ),
        "_xslt_transformer": attr.label(
            default = "@contrib_rules_jvm//java/src/com/github/bazel_contrib/contrib_rules_jvm/xml:xslt_transformer",
            executable=True,
            cfg = "host"
        ),
    },
    executable = True,
    test = True,
    doc = """Use checkstyle to lint the `srcs`.""",
)

def checkstyle_test(name, size = "medium", timeout = "short", **kwargs):
    _checkstyle_test(
        name = name,
        size = size,
        timeout = timeout,
        **kwargs
    )

CheckStyleInfo = provider(
    fields = {
        "config_file": "The config file to use.",
        "output_format": "Output Format can be plain or xml.",
    },
)

def _checkstyle_config_impl(ctx):
    return [
        DefaultInfo(
            runfiles = ctx.runfiles(ctx.files.data + [ctx.file.config_file]),
        ),
        CheckStyleInfo(
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
    },
    provides = [
        CheckStyleInfo,
    ],
    doc = """ Rule allowing checkstyle to be configured. This is typically
     used with the linting rules from `@apple_rules_lint` to configure how
     checkstyle should run. """,
)

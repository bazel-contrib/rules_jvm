PmdInfo = provider(
    fields = {
        "format": "The format to generate reports in",
        "rulesets": "Depset of files containing rulesets",
        "shallow": "Whether to use target outputs as part of processing",
    },
)

def _pmd_ruleset_impl(ctx):
    return [
        DefaultInfo(
            files = depset(ctx.files.rulesets),
        ),
        PmdInfo(
            rulesets = depset(ctx.files.rulesets),
            format = ctx.attr.format,
            shallow = ctx.attr.shallow,
        ),
    ]

pmd_ruleset = rule(
    _pmd_ruleset_impl,
    doc = "Select a rule set for PMD tests.",
    attrs = {
        "format": attr.string(
            doc = "Generate report in the given format. One of html, text, or xml (default is xml)",
            default = "xml",
            values = ["html", "text", "xml"],
        ),
        "rulesets": attr.label_list(
            doc = "Use these rulesets.",
            allow_files = True,
        ),
        "shallow": attr.bool(
            doc = "Use the targetted output to increase PMD's depth of processing",
            default = True,
        ),
    },
    provides = [
        PmdInfo,
    ],
)

load(":pmd_ruleset.bzl", "PmdInfo")

def _pmd_test_impl(ctx):
    cmd = [
        ctx.executable._pmd.short_path,
    ]

    # We want to disable the suggestion to use the analysis cache
    # https://pmd.github.io/latest/pmd_userdocs_incremental_analysis.html#disabling-incremental-analysis
    cmd.extend(["-no-cache"])

    inputs = []
    transitive_inputs = depset()

    pmd_info = ctx.attr.ruleset[PmdInfo]

    file_list = ctx.actions.declare_file("%s-pmd-srcs" % ctx.label.name)
    ctx.actions.write(
        file_list,
        ",".join([src.path for src in ctx.files.srcs]),
    )
    cmd.extend(["-filelist", file_list.short_path])
    inputs.extend(ctx.files.srcs)
    inputs.append(file_list)

    cmd.extend(["-R", ",".join([rs.path for rs in pmd_info.rulesets.to_list()])])
    inputs.extend(pmd_info.rulesets.to_list())

    cmd.extend(["-f", pmd_info.format])

    if ctx.attr.target != None and not pmd_info.shallow:
        jars = ctx.attr.target[JavaInfo].transitive_runtime_jars.to_list()
        if len(jars) > 0:
            aux_class_path = ctx.host_configuration.host_path_separator.join([jar.short_path for jar in jars])
            cmd.extend(["-auxclasspath", aux_class_path])

            # runfiles requires the depset to be in `default` order
            transitive_inputs = depset(
                direct = ctx.attr.target[JavaInfo].transitive_runtime_jars.to_list(),
                order = "default",
            )

    out = ctx.actions.declare_file(ctx.label.name)
    ctx.actions.write(out, " ".join(cmd), is_executable = True)

    runfiles = ctx.runfiles(
        files = inputs,
        transitive_files = transitive_inputs,
    )

    return DefaultInfo(
        executable = out,
        runfiles = runfiles.merge(ctx.attr._pmd[DefaultInfo].default_runfiles),
    )

pmd_test = rule(
    _pmd_test_impl,
    attrs = {
        "srcs": attr.label_list(
            allow_empty = True,
            allow_files = True,
        ),
        "target": attr.label(
            providers = [
                [JavaInfo],
            ],
        ),
        "ruleset": attr.label(
            mandatory = True,
            providers = [
                [PmdInfo],
            ],
        ),
        "_pmd": attr.label(
            cfg = "exec",
            executable = True,
            default = "//java:pmd",
        ),
    },
    executable = True,
    test = True,
    doc = """Use PMD to lint the `srcs`.""",
)

load("@apple_rules_lint//lint:defs.bzl", "LinterInfo")
load(":spotbugs_config.bzl", "SpotBugsInfo")

"""
Spotbugs integration logic
"""

def _spotbugs_impl(ctx):
    info = ctx.attr.config[SpotBugsInfo]
    max_rank = info.max_rank
    effort = info.effort
    fail_on_warning = info.fail_on_warning
    exclude_filter = info.exclude_filter
    omit_visitors = info.omit_visitors
    plugin_list = info.plugin_list

    baseline_files = ctx.attr.baseline_file
    if len(baseline_files) > 1:
        fail("More than one baseline file was specified.")

    if ctx.attr.only_output_jars:
        deps = []
        for target in ctx.attr.deps:
            if JavaInfo in target:
                # test targets do not include their own jars:
                # https://github.com/bazelbuild/bazel/issues/11705
                # work around that.
                deps.extend([jar.class_jar for jar in target[JavaInfo].outputs.jars])
        jars = depset(direct = deps).to_list()
    else:
        jars = depset(transitive = [target[JavaInfo].transitive_runtime_deps for target in ctx.attr.deps if JavaInfo in target]).to_list()

    runfiles = [info.binary]
    flags = ["-textui", "-effort:%s" % effort]
    if exclude_filter:
        flags.extend(["-exclude", exclude_filter.short_path])
        runfiles.append(exclude_filter)

    if len(baseline_files) > 0:
        baseline_file = baseline_files[0].files.to_list()[0]
        flags.extend(["-excludeBugs", baseline_file.short_path])
        runfiles.append(baseline_file)

    # NOTE: pluginList needs to be specified before specifying any visitor
    # otherwise the execution will will fail with detector not found.
    if plugin_list:
        plugin_list_cli_flag = ":".join([plugin.short_path for plugin in plugin_list])
        flags.extend(["-pluginList", plugin_list_cli_flag])
        runfiles.extend(plugin_list)

    if omit_visitors:
        flags.extend(["-omitVisitors", ",".join(omit_visitors)])

    if max_rank:
        flags.extend(["-maxRank", max_rank])

    test = [
        "#!/usr/bin/env bash",
        "ERRORLOG=$(mktemp)",
        "RES=`{lib} {flags} \"${{@}}\" {jars} 2>$ERRORLOG`".format(
            lib = info.binary.short_path,
            flags = " ".join(flags),
            jars = " ".join([jar.short_path for jar in jars]),
        ),
        "SPOTBUGS_STATUS=$?",
        "if [ $SPOTBUGS_STATUS != 0 ]; then",
        "  echo >&2 \"spotbugs exited with unexpected code $SPOTBUGS_STATUS\"",
        "  cat >&2 $ERRORLOG",
        "  exit $SPOTBUGS_STATUS",
        "fi",
        "echo \"$RES\"",
    ]

    if fail_on_warning:
        test.extend([
            "if [ -n \"$RES\" ]; then",
            "   exit 1",
            "fi",
        ])

    out = ctx.actions.declare_file(ctx.label.name + "exec")

    ctx.actions.write(
        output = out,
        content = "\n".join(test),
    )

    runfiles = ctx.runfiles(
        files = jars + runfiles,
    ).merge(ctx.attr.config[DefaultInfo].default_runfiles)

    return [
        DefaultInfo(
            executable = out,
            runfiles = runfiles,
        ),
        LinterInfo(
            language = "java",
            name = "spotbugs",
        ),
    ]

spotbugs_test = rule(
    implementation = _spotbugs_impl,
    attrs = {
        "deps": attr.label_list(
            mandatory = True,
            allow_files = False,
        ),
        "config": attr.label(
            default = "@contrib_rules_jvm//java:spotbugs-default-config",
            providers = [
                SpotBugsInfo,
            ],
        ),
        # NOTE: this will always be just one file, but we use a label_list instead of a label
        # so that we can pass a glob. In cases where no baseline file exist, the glob will simply
        # be empty. Using a label instead of a label_list would force us to create a baseline file
        # for all targets, even if theres no need for one.
        "baseline_file": attr.label_list(
            allow_files = True,
            default = [],
        ),
        "only_output_jars": attr.bool(
            doc = "If set to true, only the output jar of the target will be analyzed. Otherwise all transitive runtime dependencies will be analyzed",
            default = True,
        ),
    },
    executable = True,
    test = True,
    doc = """Use spotbugs to lint the `srcs`.""",
)

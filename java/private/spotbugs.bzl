load(":spotbugs_config.bzl", "SpotBugsInfo")
load("@apple_rules_lint//lint:defs.bzl", "LinterInfo")

"""
Spotbugs integration logic
"""

def _spotbugs_impl(ctx):
    info = ctx.attr.config[SpotBugsInfo]
    effort = info.effort
    fail_on_warning = info.fail_on_warning
    exclude_filter = info.exclude_filter

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

    test = [
        "#!/usr/bin/env bash",
        "ERRORLOG=$(mktemp)",
        "RES=`{lib} {flags} {jars} 2>$ERRORLOG`".format(
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
        "only_output_jars": attr.bool(
            doc = "If set to true, only the output jar of the target will be analyzed. Otherwise all transitive runtime dependencies will be analyzed",
            default = True,
        ),
    },
    executable = True,
    test = True,
    doc = """Use spotbugs to lint the `srcs`.""",
)

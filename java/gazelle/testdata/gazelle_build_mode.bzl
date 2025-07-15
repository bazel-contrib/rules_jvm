def _transition_impl(settings, attr):
    # Transition that enables an embedded version of the javaparser server.
    return {"//java/gazelle:embed_server": True}

bundle_prebuilt_gazelle_java_plugin = transition(
    implementation = _transition_impl,
    inputs = [],
    outputs = ["//java/gazelle:embed_server"],
)

def _gazelle_embedded_javaparser_impl(ctx):
    is_windows = ctx.configuration.host_path_separator == ";"
    gazelle = ctx.attr.gazelle_binary[0]

    java_runtime = ctx.toolchains["@bazel_tools//tools/jdk:toolchain_type"].java.java_runtime
    java_home_path = java_runtime.java_home_runfiles_path
    if is_windows:
        java_home_path.replace("/", "\\")

    gazelle_bin = gazelle[DefaultInfo].files_to_run.executable

    gazelle_runner_script_name = ctx.attr.name + (".bat" if is_windows else "")
    gazelle_runner = ctx.actions.declare_file(gazelle_runner_script_name)

    gazelle_runner_script_unix = """#!/bin/bash
# Because we execute the gazelle binary directly, we can't rely that CWD will be the Bazel sandbox root.
# Therefore, we have to look in the runfiles dir for the real gazelle binary.
mkdir -p $RUNFILES_DIR/hack_for_runtime_path
export GAZELLE_JAVA_JAVAHOME="$RUNFILES_DIR/hack_for_runtime_path/{java_home}"

# If both the embedded and non-embedded version of a test are running at the same time, they will collide trying to create the port file and crash.
export TMPDIR="$TMPDIR_embedded"

exec $RUNFILES_DIR/_main/{gazelle_bin} "$@"
    """

    gazelle_runner_script_windows = """@echo off
REM Because java_home points to ../<java_repository> but we pass the JDK runfiles in DefaultInfo,
REM we insert a non-existent directory to cancel out the `..`.
set "GAZELLE_JAVA_JAVAHOME=%RUNFILES_DIR%\\hack_for_runtime_path\\{java_home}"

REM If both the embedded and non-embedded version of a test are running at the same time, they will collide trying to create the port file and crash.
if not exist %TMPDIR%\\_embedded\\NUL mkdir "%TMPDIR%\\_embedded"
set "TMPDIR=%TMPDIR%\\_embedded"

call %RUNFILES_DIR%\\_main\\{gazelle_bin} %*
"""

    gazelle_bin_path = gazelle_bin.short_path

    gazelle_runner_script = gazelle_runner_script_windows if is_windows else gazelle_runner_script_unix

    ctx.actions.write(
        output = gazelle_runner,
        content = gazelle_runner_script.format(
            java_home = java_home_path,
            gazelle_bin = gazelle_bin_path,
        ),
        is_executable = True,
    )

    return [
        DefaultInfo(
            executable = gazelle_runner,
            runfiles = ctx.runfiles(files = [gazelle_bin], transitive_files = java_runtime.files),
            files = depset([gazelle_runner]),
        ),
    ]

gazelle_embedded_javaparser = rule(
    doc = "Rule that transitions gazelle to embed the javaparser server jar into the gazelle binary",
    implementation = _gazelle_embedded_javaparser_impl,
    attrs = {
        "gazelle_binary": attr.label(
            executable = True,
            cfg = bundle_prebuilt_gazelle_java_plugin,
        ),
    },
    toolchains = ["@bazel_tools//tools/jdk:toolchain_type"],
    executable = True,
)

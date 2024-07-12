#!/usr/bin/env python3

import argparse
import json
import os
import re
import subprocess
import sys
import textwrap
import zipfile
from os import path
from pathlib import Path

UNCHANGED_FILES = ["outdated.sh", "outdated.artifacts", "outdated.repositories", "compat_repository.bzl"]

parser = argparse.ArgumentParser(
    description="""Convert rules_jvm_external lock files to frozen zip files that can be 
imported in your own builds.""",
    formatter_class=argparse.ArgumentDefaultsHelpFormatter,
)
parser.add_argument(
    "--repo",
    default="contrib_rules_jvm_deps",
    help="Name of the `maven_install` rule to freeze",
)
parser.add_argument(
    "--zip",
    default="java/private/contrib_rules_jvm_deps.zip",
    help="Name of zip file to create containing frozen deps"
)

parser.add_argument(
    "--zip-repo",
    help="Name of the zip repository used to import the zip file. Used only if compat_repositories are enabled"
)

# When run under `bazel run` this directory will be set. If not,
# then we should assume our current directory is the right one
# to use.
cwd = os.environ.get("BUILD_WORKSPACE_DIRECTORY", None)

args = parser.parse_args()

pin_env = {"REPIN": "1"}
pin_env.update(os.environ)

# Repin our dependencies
cmd = ["bazel", "run", "@%s//:pin" % args.repo]
subprocess.check_call(cmd, env=pin_env, cwd = cwd)

# Now grab the files we need from their output locations
cmd = ["bazel", "info", "output_base"]
output_base = subprocess.check_output(cmd, cwd = cwd).rstrip()
base = output_base.decode(encoding=sys.stdin.encoding)

# Figure out the mangled repo name
cmd = ["bazel", "cquery", "--output=starlark", "--starlark:expr=target.label.workspace_name", "@{name}//:outdated".format(name = args.repo)]
base_dir = subprocess.check_output(cmd, cwd=cwd, stderr=subprocess.DEVNULL).decode('utf-8').strip()

root = Path(base) / "external" / base_dir

# Generate a stable-ish zip file
zip_path = path.join(cwd, args.zip) if cwd else args.zip
output = zipfile.ZipFile(zip_path, "w", zipfile.ZIP_DEFLATED)

for f in UNCHANGED_FILES:
    p = root / f
    zinfo = zipfile.ZipInfo(filename=f, date_time=(1980, 1, 1, 0, 0, 0))
    if path.exists(p):
        with open(p) as input:
            output.writestr(zinfo, input.read())

defs_bzl = root / "defs.bzl"
zinfo = zipfile.ZipInfo(filename='defs.bzl', date_time=(1980, 1, 1, 0, 0, 0))
with open(defs_bzl) as f:
    # We need to strip the `netrc` lines from this file
    defs = []
    for line in f.read().split('\n'):
        line = re.sub(r'^\s+netrc\s+=\s+.*$', '', line)
        if len(line):
            defs.append(line)
    output.writestr(zinfo, "\n".join(defs))

# Generate a compat.bzl file. This is only needed for workspace-based builds, but the
# file is not generated in a `bzlmod`-created `@maven`.
lock_file = root / "imported_maven_install.json"

compat_content = """load("@contrib_rules_jvm_deps//:compat_repository.bzl", "compat_repository")

def compat_repositories():
    """

with open(lock_file) as lf:
    parsed_lock_file = json.load(lf)
    if parsed_lock_file["version"] != "2":
        raise Exception("Lock file needs to be version 2")
    for name in parsed_lock_file["artifacts"].keys():
        munged = name.replace(".", "_").replace("-", "_").replace(":", "_")
        compat_content += f"""
    compat_repository(
        name = "{munged}",
        generating_repository = "{args.repo}",
    )
"""
zinfo = zipfile.ZipInfo(filename="compat.bzl", date_time=(1980, 1, 1, 0, 0, 0))
# All files are readable by everyone
zinfo.external_attr = 0o666 << 16
output.writestr(zinfo, compat_content)

build_file = root / "BUILD"
zinfo = zipfile.ZipInfo(filename='BUILD.bazel', date_time=(1980, 1, 1, 0, 0, 0))
with open(build_file) as f:
    build_file_contents = textwrap.dedent(
        """\
        load("@bazel_skylib//:bzl_library.bzl", "bzl_library")
        
        {original_contents}
        """.format(original_contents = textwrap.indent(f.read(), "        "))
    )
    output.writestr(zinfo, build_file_contents)

output.close()

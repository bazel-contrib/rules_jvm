#!/usr/bin/env python3

import argparse
import os
import re
import subprocess
import sys
import textwrap
import zipfile
from os import path

UNCHANGED_FILES = ["outdated.sh", "outdated.artifacts", "outdated.repositories", "compat_repository.bzl"]

parser = argparse.ArgumentParser(
    description="""Convert rules_jvm_external lock files to frozen zip files that can be 
imported in your own builds.""",
    formatter_class=argparse.ArgumentDefaultsHelpFormatter,
)
parser.add_argument(
    "--repo",
    default="frozen_deps",
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

args = parser.parse_args()

pin_env = {"REPIN": "1"}
pin_env.update(os.environ)

# Repin our dependencies
cmd = ["bazel", "run", "@unpinned_%s//:pin" % args.repo]
subprocess.check_call(cmd, env=pin_env)

# Now grab the files we need from their output locations
cmd = ["bazel", "info", "output_base"]
output_base = subprocess.check_output(cmd).rstrip()
base = output_base.decode(encoding=sys.stdin.encoding)

# Generate a stable-ish zip file
output = zipfile.ZipFile(args.zip, "w", zipfile.ZIP_DEFLATED)

for f in UNCHANGED_FILES:
    p = path.join(base, "external", args.repo, f)
    zinfo = zipfile.ZipInfo(filename=f, date_time=(1980, 1, 1, 0, 0, 0))
    if path.exists(p):
        with open(p) as input:
            output.writestr(zinfo, input.read())

defs_bzl = path.join(base, "external", args.repo, "defs.bzl")
zinfo = zipfile.ZipInfo(filename='defs.bzl', date_time=(1980, 1, 1, 0, 0, 0))
with open(defs_bzl) as f:
    # We need to strip the `netrc` lines from this file
    defs = []
    for line in f.read().split('\n'):
        line = re.sub(r'^\s+netrc\s+=\s+.*$', '', line)
        if len(line):
            defs.append(line)
    output.writestr(zinfo, "\n".join(defs))

# Copy the compat.bzl if it was added to the maven install via generate_compat_repositories = True attribute
# need to update the repo name in the file.
compat_bzl = path.join(base, "external", args.repo, "compat.bzl")
if path.exists(compat_bzl):
    zinfo = zipfile.ZipInfo(filename="compat.bzl", date_time=(1980, 1, 1, 0, 0, 0))
    libname = args.zip_repo if args.zip_repo else path.basename(path.splitext(args.zip)[0])
    with open(compat_bzl) as f:
        lines = [re.sub(args.repo, libname, line) for line in f.read().split('\n')]
    output.writestr(zinfo, "\n".join(lines))


build_file = path.join(base, "external", args.repo, "BUILD")
zinfo = zipfile.ZipInfo(filename='BUILD.bazel', date_time=(1980, 1, 1, 0, 0, 0))
with open(build_file) as f:
    build_file_contents = textwrap.dedent(
        """\
        load("@bazel_skylib//:bzl_library.bzl", "bzl_library")
        
        {original_contents}
        
        bzl_library(
           name = "defs",
           srcs = ["defs.bzl"],
           deps = [
               "@rules_jvm_external//:implementation",
           ],
        )
        """.format(original_contents = textwrap.indent(f.read(), "        "))
    )
    output.writestr(zinfo, build_file_contents)

output.close()

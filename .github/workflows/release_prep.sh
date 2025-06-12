#!/usr/bin/env bash
set -eufo pipefail

# Called by .github/workflows/release.yml to generate release notes.

# Set by GH actions, see
# https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
TAG=${GITHUB_REF_NAME} # e.g. v1.2.3
VERSION=${TAG#v}       # e.g. 1.2.3
# The prefix is chosen to match what GitHub generates for source archives
PREFIX="rules_jvm-${TAG:1}"
ARCHIVE="rules_jvm-$TAG.tar.gz"
git archive --format=tar --prefix="${PREFIX}/" "${TAG}" | gzip >"$ARCHIVE"

cat <<EOF
\`contrib_rules_jvm\` only supports \`bzlmod\`-enabled builds

## Module setup

In your \`MODULE.bazel\`:

\`\`\`starlark
bazel_dep(name = "contrib_rules_jvm", version = "${VERSION}")
\`\`\`
EOF

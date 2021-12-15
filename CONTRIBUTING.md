# How to Contribute

## Formatting

Starlark files should be formatted by buildifier.

`buildifier --mode fix -lint fix -r .`

## Using this as a development dependency of other rules

You'll commonly find that you develop in another WORKSPACE, such as
some other ruleset that depends on `contrib_rules_jvm`, or in a nested
WORKSPACE in the integration_tests folder.

To always tell Bazel to use this directory rather than some release
artifact or a version fetched from the internet, run this from this
directory:

```sh
OVERRIDE="--override_repository=contrib_rules_jvm=$(pwd)/rules_jvm"
echo "build $OVERRIDE" >> ~/.bazelrc.user
echo "fetch $OVERRIDE" >> ~/.bazelrc.user
echo "query $OVERRIDE" >> ~/.bazelrc.user
```

This means that any usage of `@contrib_rules_jvm` on your system will
point to the `rules_jvm` folder.

## Releasing

1. Determine the next release version, following semver if possible/
2. Tag the repo and push it (or create a tag in GH UI).
3. Watch the automation run on GitHub actions.


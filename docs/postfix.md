## Freezing Dependencies

At runtime, a handful of dependencies are required by helper classes
in this project. Rather than pollute the default `@maven` workspace,
these are loaded into a `@contrib_rules_jvm_deps` workspace. These
dependencies are loaded using a call to `maven_install`, but we don't
want to force users to remember to load our own dependencies for
us. Instead, to add a new dependency to the project:

1. Update `frozen_deps` in the `WORKSPACE` file
2. Run `./bin/freeze-deps.py`
3. Commit the updated files.

### Freezing your dependencies

As noted above, if you are building Bazel rules which require Java
parts, and hence Java dependencies, it can be useful to freeze these
dependencies. This process captures the list of dependencies into a
zip file.  This zip file is distributed with your Bazel rules. This
makes your rule more hermetic, your rule no longer relies on the user
to supply the correct dependencies, because they get resolved under
their own repository namespace rather than being intermingled with the
user's, and your dependencies no longer conflict with the users
selections. If you would like to create your dependencies as a frozen
file you need to do the following:

1. Create a maven install rule with your dependencies and with a
   unique name, this should be in or referenced by your WORKSPACE
   file.
   ```starlark
   maven_install(
        name = "frozen_deps",
        artifacts = [...],
        fail_if_repin_required = True,
        fetch_sources = True,
        maven_install_json = "@workspace//:frozen_deps_install.json",
   )
   
   load("@frozen_deps//:defs.bzl", "pinned_maven_install")
   
   pinned_maven_install()
   ```
2. Run `./bin/frozen-deps.py --repo <repo name> --zip
   <path/to/dependency.zip>`. The `<repo name>` matches the name used
   for the `maven_install()` rule above. This will pin the
   dependencies then collect them into the zip file.
3. Commit the zip file into your repo.
4. Add a `zip_repository()` rule to your WORKSPACE to configure the
   frozen dependencies:
   ```starlark
   maybe(
       zip_repository,
       name = "workspace_deps",
       path = "@workspace//path/to:dependency.zip",
   )
   ```
5. Make sure to pin the maven install from the repostitory:
   ```starlark
   load("@workspace_deps//:defs.bzl", "pinned_maven_install")
   
   pinned_maven_install()
   ```
6. Use the dependencies for your `java_library` rules from the frozen
   deps:
   ```starlark
   java_library(
        name = "my_library",
        srcs = glob(["*.java"]),
        deps = [
             "@workspace_deps//:my_frozen_dep",
        ],
   )
   ```

NOTE: If you need to generate the compat_repositories for the
dependencies, usually because your rule depends on another rule which
is still using the older compat repositories, you need to make the
following changes and abide by the following restrictions.

1. Add the `generate_compat_repositories = True,` attribute to the
   original `maven_install()` rule.
2. In step 2, add the parameter `--zip-repo workspace_deps` to match
   the name used in the `zip_repository()` rule (step 4). If you don't
   supply this, it uses the base name of the zip file, which may not
   be what you want.
3. In step 5, add the call to generate the compat_repositories:
   ```starlark
   load("@workspace_deps//:compat.bzl", "compat_repositories")
   
   compat_repositories()
   ```

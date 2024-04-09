# Invalid label in runtime_deps

The extension loads the existing runtime_deps to reuse and add to them.

This mandates that the runtime_deps is a valid string list of labels. Invalid
labels will cause the extension to exit with 1.

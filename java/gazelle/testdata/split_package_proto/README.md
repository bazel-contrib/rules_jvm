# Split Package - Proto Test

This test verifies that Gazelle correctly resolves dependencies when a Java source file
references a proto-generated class from the same Java package but a different Bazel package.

## Scenario

- `proto/src/main/proto/com/example/` contains `options.proto` with `java_package="com.example"`
- `java/src/main/java/com/example/` contains `Client.java` in package `com.example`
- `Client.java` uses `Options.ResponseCode` without an import (same Java package)

## Key Points

1. The proto file specifies `java_outer_classname = "Options"`, so the generated class is `com.example.Options`
2. Java code references `Options.ResponseCode` - this is an outer class with an inner enum
3. Since both are in the same Java package, no import is needed in Java
4. Gazelle must detect `Options` as a same-package type reference and resolve it to the proto target

## Expected Behavior

The generated BUILD for `java/src/main/java/com/example/` should include:
```
deps = ["//proto/src/main/proto/com/example:example_java_library"]
```

## What This Tests

1. The Java parser correctly identifies `Options` from `Options.ResponseCode` as a same-package reference
2. The proto extension indexes the `java_library` with the correct `java_package`
3. The resolver finds the proto target when multiple providers exist for the same package

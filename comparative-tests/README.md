# Comparing outputs from test runners

This directory contains test classes that build and test using a variety of 
different Java build tools. This is helpful if you want to check comparative 
behaviours between these tools.

To run with Maven (assuming you're in the `comparative-tests` directory):

`mvn --fail-at-end test`

With Gradle (assuming you're in the `comparative-tests` directory):

`./gradlew test`

With Bazel:

`bazel test //comparative-tests/...`

Tests logs may be found:

* Maven: `comparative-tests/target/surefire-reports`
* Gradle: `comparative-tests/build/test-results/test`
* Bazel: `bazel-testlogs/comparative-tests/src/test/java/com/apple/sdp/gradle/`

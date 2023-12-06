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


## Updating `gradlew`

Gradle projects typically ship with a `gradlew` script, and this directory 
is no exception. By doing this, we avoid the need to make users install 
`gradle` on their systems, but it does mean that we have seemingly random 
files scattered around.

To [update `gradlew`][gradlew] run the command: `./gradlew wrapper 
--gradle-version latest`

[gradlew]: https://docs.gradle.org/current/userguide/gradle_wrapper.html#sec:upgrading_wrapper
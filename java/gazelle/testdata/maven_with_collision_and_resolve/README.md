# Maven with collision and resolve

Make sure the java extension does not fail on colliding packages in maven if the correct resolve directives are present.

Note that the maven_install.json file is manually crafted/invalid in order to simulate a collision on the `com.google.common.primitives` import.
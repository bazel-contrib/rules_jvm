package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import org.junit.jupiter.api.Test;

import java.io.File;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.zip.ZipFile;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.fail;

public class JavaTestSuiteTest {

    @Test
    public void testClasspathDoesNotContainDuplicateResources() throws IOException {
        final String resourceName = "duplicate_resource_test.txt";
        String[] classpath = System.getProperty("java.class.path").split(":");
        List<String> resourcesFound = new ArrayList<>();
        for (String entry : classpath) {
            if (entry.endsWith(".jar")) {
                try (ZipFile zipfile = new ZipFile(new File(entry))) {
                    zipfile.entries().asIterator().forEachRemaining(zipEntry -> {
                        if (zipEntry.getName().endsWith(resourceName)) {
                            resourcesFound.add(entry);
                        }
                    });
                }
            } else {
                fail("Classpath contains non-jar files");
            }
        }
        assertEquals(1, resourcesFound.size(),
                "Expected a single instance of a resource to be on the classpath, instead found entries in: " + resourcesFound);
    }
}

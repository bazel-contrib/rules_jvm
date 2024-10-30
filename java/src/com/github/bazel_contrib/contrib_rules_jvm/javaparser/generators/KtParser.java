package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import org.jetbrains.kotlin.parsing.KotlinParser;
import java.nio.file.Path;
import java.util.List;

import org.jetbrains.kotlin.parsing.KotlinParser;

public class KtParser {
    private final KotlinParser parser = new KotlinParser(null);

    public ParsedPackageData parseClasses(Path directory, List<String> files) {
        ParsedPackageData result = new ParsedPackageData();
        result.packages.add("com.example");
        // TODO: Implement something more robust.
        return result;
    }
}
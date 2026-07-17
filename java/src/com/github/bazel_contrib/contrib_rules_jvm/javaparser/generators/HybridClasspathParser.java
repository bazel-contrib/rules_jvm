package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Parses Java source files using Turbine for declaration-level analysis, with an automatic
 * per-file fallback to javac for files that contain {@code .class} literals unresolved by Turbine.
 *
 * <p>Turbine strips method bodies, so same-package types used only via {@code Foo.class} literals
 * (without an explicit import) are not captured. This matters for split-package targets where
 * {@code Foo} may live in a different Bazel target. The spot-check after the Turbine pass detects
 * exactly those files and re-parses them with javac, keeping the fast path for the rest.
 *
 * <p>A {@code # gazelle:java-parser full} directive can force javac for an entire directory when
 * the auto-detection is insufficient.
 */
public class HybridClasspathParser {

  private static final Logger logger = LoggerFactory.getLogger(HybridClasspathParser.class);

  // Matches PascalCase.class expressions (e.g. MyHandler.class, String.class)
  private static final Pattern CLASS_LITERAL = Pattern.compile("\\b([A-Z]\\w+)\\.class\\b");

  private final ClasspathParser javacParser = new ClasspathParser();

  public ParsedPackageData parseClasses(Path directory, List<String> files) throws IOException {
    // Step 1: Read each file once; parse with Turbine and cache the source text
    Map<String, String> sources = new LinkedHashMap<>();
    ParsedPackageData turbineResult = new ParsedPackageData();

    for (String filename : files) {
      Path path = directory.resolve(filename);
      try {
        String source = Files.readString(path);
        sources.put(filename, source);
        turbineResult.merge(TurbineClasspathParser.parseOneFile(path, source));
      } catch (IOException e) {
        logger.warn("HybridClasspathParser: cannot read {}", path);
      }
    }

    // Step 2: Build the set of simple class names already resolved by Turbine across this
    // directory. If a class is already in usedTypes (via an explicit import in any file), any
    // .class literal for it is already covered — no fallback needed.
    Set<String> resolvedSimpleNames = new HashSet<>();
    for (String fqn : turbineResult.usedTypes) {
      resolvedSimpleNames.add(simpleName(fqn));
    }

    // Step 3: Identify files that have .class literals for unresolved (same-package) types
    List<String> needsJavac = new ArrayList<>();
    for (Map.Entry<String, String> entry : sources.entrySet()) {
      if (hasUnresolvedClassLiteral(entry.getValue(), resolvedSimpleNames)) {
        needsJavac.add(entry.getKey());
      }
    }

    // Step 4: Run javac only on the flagged files and merge the additional types found
    if (!needsJavac.isEmpty()) {
      try {
        turbineResult.merge(javacParser.parseClasses(directory, needsJavac));
      } catch (Exception e) {
        logger.error("HybridClasspathParser: javac fallback failed for {}", directory, e);
      }
    }

    return turbineResult;
  }

  private static boolean hasUnresolvedClassLiteral(String source, Set<String> resolvedSimpleNames) {
    Matcher m = CLASS_LITERAL.matcher(source);
    while (m.find()) {
      if (!resolvedSimpleNames.contains(m.group(1))) {
        return true;
      }
    }
    return false;
  }

  private static String simpleName(String fqn) {
    int lastDot = fqn.lastIndexOf('.');
    return lastDot >= 0 ? fqn.substring(lastDot + 1) : fqn;
  }
}

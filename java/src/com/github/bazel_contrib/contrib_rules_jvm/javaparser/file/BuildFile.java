package com.github.bazel_contrib.contrib_rules_jvm.javaparser.file;

import static java.nio.file.StandardOpenOption.TRUNCATE_EXISTING;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.regex.Pattern;
import javax.annotation.Nullable;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class BuildFile {
  private static final Logger logger = LoggerFactory.getLogger(BuildFile.class);

  /**
   * Full path the the build file e.g.
   * /root/workspace/project/src/main/com/gazelle/java/javaparser/BUILD.bazel
   */
  private final Path buildPath;

  /**
   * Contents of the BUILD file we're working with. Read on class creation, updated and written on
   * command
   */
  @Nullable private String content;

  /**
   * Name of the Target in the build file to be updating, Name of the dependencies list in the build
   * file contents to be updated Default follows the Bazel target rules and names it after the tail
   * directory name. e.g. file for this package.
   */
  private final String targetName;

  /** Should be the target for bazel. eg. //src/main/com/gazelle/java/javaparser/file */
  private final String bazelTarget;

  public BuildFile(Path buildFile) {
    this(buildFile, null, null);
    this.content = BuildFile.parse(buildPath);
  }

  public BuildFile(Path buildFile, String bazelTarget, String content) {
    this.buildPath = buildFile;
    this.targetName = buildFile.getParent().getFileName().toString();
    this.bazelTarget = bazelTarget;
    this.content = content;
  }

  public String getTargetName() {
    return targetName;
  }

  public Path getBuildPath() {
    return buildPath;
  }

  public String getBazelTarget() {
    return bazelTarget;
  }

  public String replaceTarget(String target, String replace) {
    String match = targetName + target + " = \\[.*?\\]\n\n";
    Pattern pattern = Pattern.compile(match, Pattern.DOTALL);
    return pattern.matcher(content).replaceFirst(replace);
  }

  public void writeUpdatedFile(String update) throws IOException {
    Files.writeString(buildPath, update, TRUNCATE_EXISTING);
  }

  public static String parse(Path buildPath) {
    try {
      return Files.readString(buildPath);
    } catch (IOException ex) {
      logger.error("Unable to read build file: {} : {}", buildPath, ex);
    }
    return null;
  }
}

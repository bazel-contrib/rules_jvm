package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import org.apache.commons.cli.CommandLine;
import org.apache.commons.cli.CommandLineParser;
import org.apache.commons.cli.DefaultParser;
import org.apache.commons.cli.HelpFormatter;
import org.apache.commons.cli.Option;
import org.apache.commons.cli.Options;
import org.apache.commons.cli.ParseException;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Main {
  private static final Logger logger = LoggerFactory.getLogger(Main.class);
  private static CommandLine line;

  public static void main(String[] args) throws IOException, InterruptedException {
    line = commandLineOptions(args);

    Main main = new Main();
    if (line.hasOption("server")) {
      main.runServer();
    } else {
      main.run();
    }
  }

  public void runServer() throws InterruptedException, IOException {
    PackageParser parser = new PackageParser(workspace());
    parser.setup(srcs(), tests(), generated());
    GrpcServer gRPCServer = new GrpcServer(serverPort(), parser);
    gRPCServer.start();
    gRPCServer.blockUntilShutdown();
  }

  public void run() throws IOException {
    PackageParser parser = new PackageParser(workspace());
    parser.setup(srcs(), tests(), generated());
    if (imports() != null) {
      parser.runImports(imports());
    } else {
      parser.runAll(dryRun());
    }
  }

  private Path workspace() {
    return line.hasOption("workspace")
        ? Paths.get(line.getOptionValue("workspace"))
        : Paths.get("");
  }

  private boolean dryRun() {
    return line.hasOption("dry-run");
  }

  private String srcs() {
    return line.hasOption("sources") ? line.getOptionValue("sources") : "**/src/{main/java,main}";
  }

  private String tests() {
    return line.hasOption("tests") ? line.getOptionValue("tests") : "**/src/{test/java,test}";
  }

  private String generated() {
    return line.getOptionValue("generated");
  }

  private String imports() {
    return line.getOptionValue("parsed-imports");
  }

  private int serverPort() {
    return line.hasOption("server-port")
        ? Integer.decode(line.getOptionValue("server-port"))
        : 8980;
  }

  private static CommandLine commandLineOptions(String[] args) {
    Options options = new Options();
    CommandLineParser parser = new DefaultParser();

    options.addOption(new Option("h", "help", false, "This help message"));
    options.addOption(
        new Option(
            null,
            "server",
            false,
            "Start the Build File gRPC server to manage multiple groups of files"));
    options.addOption(
        new Option(
            null, "dry-run", false, "Output only, but do not change files in the workspace"));

    options.addOption(new Option(null, "workspace", true, "Workspace root"));
    options.addOption(
        new Option(null, "server-port", true, "Port to connect to the gRPC server (default 8980)"));
    options.addOption(new Option(null, "sources", true, "Relative path to java sources"));
    options.addOption(new Option(null, "tests", true, "Relative path to java tests"));
    options.addOption(new Option(null, "generated", true, "Relative path to generated code"));
    options.addOption(
        new Option(
            null,
            "parse-imports",
            true,
            "Generate the imports from list of files (use wild cards)"));

    try {
      line = parser.parse(options, args);
    } catch (ParseException e) {
      logger.error("Command line parsing failed. {}", e.getMessage());
      System.exit(3);
    }
    // Display the help if requested.
    if (line.hasOption("help")) {
      HelpFormatter formatter = new HelpFormatter();
      formatter.printHelp("build-file-generator", options);
      System.exit(0);
    }
    return line;
  }
}

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
    main.runServer();
  }

  public void runServer() throws InterruptedException, IOException {
    GrpcServer gRPCServer = new GrpcServer(serverPort(), workspace());
    gRPCServer.start();
    gRPCServer.blockUntilShutdown();
  }

  private Path workspace() {
    return line.hasOption("workspace")
        ? Paths.get(line.getOptionValue("workspace"))
        : Paths.get("");
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
    options.addOption(new Option(null, "workspace", true, "Workspace root"));
    options.addOption(
        new Option(null, "server-port", true, "Port to connect to the gRPC server (default 8980)"));

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

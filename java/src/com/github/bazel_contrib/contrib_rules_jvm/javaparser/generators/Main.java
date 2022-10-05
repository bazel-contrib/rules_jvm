package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.concurrent.ScheduledThreadPoolExecutor;
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

    TimeoutHandler timeoutHander =
        new TimeoutHandler(new ScheduledThreadPoolExecutor(1), main.idleTimeout());
    main.runServer(timeoutHander);
  }

  public void runServer(TimeoutHandler timeoutHandler) throws InterruptedException, IOException {
    GrpcServer gRPCServer = new GrpcServer(serverPortFilePath(), workspace(), timeoutHandler);
    gRPCServer.start();
    gRPCServer.blockUntilShutdown();
  }

  private Path workspace() {
    return line.hasOption("workspace")
        ? Paths.get(line.getOptionValue("workspace"))
        : Paths.get("");
  }

  private Path serverPortFilePath() {
    return Paths.get(line.getOptionValue("server-port-file-path"));
  }

  // <=0 means don't timeout.
  private int idleTimeout() {
    return line.hasOption("idle-timeout")
        ? Integer.decode(line.getOptionValue("idle-timeout"))
        : -1;
  }

  private static CommandLine commandLineOptions(String[] args) {
    Options options = new Options();
    CommandLineParser parser = new DefaultParser();

    options.addOption(new Option("h", "help", false, "This help message"));
    options.addOption(new Option(null, "workspace", true, "Workspace root"));
    options.addOption(new Option(null, "server-port-file-path", true, "TODO"));
    options.addOption(
        new Option(
            null,
            "idle-timeout",
            true,
            "Number of seconds after no gRPC activity to terminate self"));

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

package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import com.gazelle.java.javaparser.v0.JavaParserGrpc;
import com.gazelle.java.javaparser.v0.Package;
import com.gazelle.java.javaparser.v0.ParsePackageRequest;
import com.google.common.collect.Iterables;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.protobuf.services.ProtoReflectionService;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;
import java.util.concurrent.TimeUnit;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class GrpcServer {
  private static final Logger logger = LoggerFactory.getLogger(GrpcServer.class);

  private final int port;
  private final Server server;

  /** Create a BuildFileGenerator server using serverBuilder as a base and features as data. */
  public GrpcServer(int port, Path workspace) {
    this.port = port;
    ServerBuilder serverBuilder = ServerBuilder.forPort(port);
    this.server =
        serverBuilder
            .addService(new GrpcService(workspace))
            .addService(ProtoReflectionService.newInstance())
            .build();
  }

  /** Start serving requests. */
  public void start() throws IOException {
    server.start();
    logger.debug("Server started, listening on {}", port);
    Runtime.getRuntime()
        .addShutdownHook(
            new Thread() {
              @Override
              public void run() {
                // Use stderr here since the logger may have been reset by its JVM
                // shutdown hook.
                System.err.println("*** shutting down gRPC server since JVM is shutting down");
                try {
                  GrpcServer.this.stop();
                } catch (InterruptedException e) {
                  e.printStackTrace(System.err);
                }
                System.err.println("*** server shut down");
              }
            });
  }

  /** Stop serving requests and shutdown resources. */
  public void stop() throws InterruptedException {
    server.shutdown().awaitTermination(30, TimeUnit.SECONDS);
  }

  /** Await termination on the main thread since the grpc library uses daemon threads. */
  public void blockUntilShutdown() throws InterruptedException {
    server.awaitTermination();
  }

  private static class GrpcService extends JavaParserGrpc.JavaParserImplBase {

    private final Path workspace;

    GrpcService(Path workspace) {
      this.workspace = workspace;
    }

    @Override
    public void parsePackage(
        ParsePackageRequest request, StreamObserver<Package> responseObserver) {
      logger.debug("Got request, now processing");
      try {
        responseObserver.onNext(getImports(request));
      } catch (Exception ex) {
        logger.error("Got Exception parsing package {}: {}", Paths.get(request.getRel()), ex.getMessage());
        responseObserver.onError(ex);
      }
      responseObserver.onCompleted();
      logger.debug("Finished processing request");
    }

    private Package getImports(ParsePackageRequest request) {
      List<String> files = new ArrayList<>();
      for (int i = 0; i < request.getFilesCount(); i++) {
        files.add(request.getFiles(i));
      }
      logger.debug("Working relative directory: {}", request.getRel());
      logger.debug("processing files: {}", files);

      ClasspathParser parser = new ClasspathParser();
      Path directory = workspace.resolve(request.getRel());

      try {
        parser.parseClasses(directory, files);
      } catch (IOException exception) {
        // If we fail to process a directory, which can happen with the module level processing
        // or can't parse any of the files, just return an empty response.
        return Package.newBuilder()
                .setName("")
                .build();
      }
      Set<String> packages = parser.getPackages();
      if (packages.size() > 1) {
        logger.error(
            "Set of classes in {} should have only one package, instead is: {}",
            request.getRel(),
            packages);
        throw new StatusRuntimeException(Status.INVALID_ARGUMENT);
      } else if (packages.isEmpty()) {
        logger.info("Set of classes in {} has no package",Paths.get(request.getRel()).toAbsolutePath());
        packages.add("");
      }
      logger.debug("Got package: {}", Iterables.getOnlyElement(packages));
      logger.debug("Got used types: {}", parser.getUsedTypes());
      return Package.newBuilder()
          .setName(Iterables.getOnlyElement(packages))
          .addAllImports(parser.getUsedTypes())
          .addAllMains(parser.getMainClasses())
          .build();
    }
  }
}

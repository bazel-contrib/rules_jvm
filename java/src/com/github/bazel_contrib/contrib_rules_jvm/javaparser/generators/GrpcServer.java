package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import static java.nio.file.StandardCopyOption.ATOMIC_MOVE;

import com.gazelle.java.javaparser.v0.JavaParserGrpc;
import com.gazelle.java.javaparser.v0.Package;
import com.gazelle.java.javaparser.v0.Package.Builder;
import com.gazelle.java.javaparser.v0.ParsePackageRequest;
import com.gazelle.java.javaparser.v0.PerClassMetadata;
import com.google.common.collect.ImmutableSet;
import com.google.common.collect.ImmutableSortedSet;
import com.google.common.collect.Iterables;
import io.grpc.Server;
import io.grpc.ServerBuilder;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.protobuf.services.ProtoReflectionService;
import io.grpc.stub.StreamObserver;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.concurrent.TimeUnit;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class GrpcServer {
  private static final Logger logger = LoggerFactory.getLogger(GrpcServer.class);

  private final Path serverPortFilePath;
  private final TimeoutHandler timeoutHandler;
  private final Server server;

  /** Create a BuildFileGenerator server using serverBuilder as a base and features as data. */
  public GrpcServer(Path serverPortFilePath, Path workspace, TimeoutHandler timeoutHandler) {
    this.serverPortFilePath = serverPortFilePath;
    this.timeoutHandler = timeoutHandler;
    ServerBuilder serverBuilder = ServerBuilder.forPort(0);
    this.server =
        serverBuilder
            .addService(new GrpcService(workspace, timeoutHandler))
            .addService(new LifecycleService())
            .addService(ProtoReflectionService.newInstance())
            .build();
  }

  /** Start serving requests. */
  public void start() throws IOException {
    server.start();

    // Atomically write our server port to a file so that a reading process can't do a partial read.
    Path tmpPath = serverPortFilePath.resolveSibling(serverPortFilePath.getFileName() + ".tmp");
    Files.write(tmpPath, String.format("%d", server.getPort()).getBytes(StandardCharsets.UTF_8));
    Files.move(tmpPath, serverPortFilePath, ATOMIC_MOVE);

    logger.debug("Server started, listening on {}", server.getPort());
    Runtime.getRuntime()
        .addShutdownHook(
            new Thread() {
              @Override
              public void run() {
                timeoutHandler.cancelOutstandingAndStopScheduling();
                try {
                  GrpcServer.this.stop();
                } catch (InterruptedException e) {
                  e.printStackTrace(System.err);
                }
              }
            });
  }

  /** Stop serving requests and shutdown resources. */
  public void stop() throws InterruptedException {
    server.shutdownNow().awaitTermination(30, TimeUnit.SECONDS);
  }

  /** Await termination on the main thread since the grpc library uses daemon threads. */
  public void blockUntilShutdown() throws InterruptedException {
    server.awaitTermination();
  }

  private static class GrpcService extends JavaParserGrpc.JavaParserImplBase {

    private final Path workspace;
    private final TimeoutHandler timeoutHandler;

    GrpcService(Path workspace, TimeoutHandler timeoutHandler) {
      this.workspace = workspace;
      this.timeoutHandler = timeoutHandler;
    }

    @Override
    public void parsePackage(
        ParsePackageRequest request, StreamObserver<Package> responseObserver) {
      timeoutHandler.startedRequest();

      try {
        responseObserver.onNext(getImports(request));
        responseObserver.onCompleted();
      } catch (Exception ex) {
        logger.error(
            "Got Exception parsing package {}: {}", Paths.get(request.getRel()), ex.getMessage());
        responseObserver.onError(ex);
        responseObserver.onCompleted();
      } finally {
        timeoutHandler.finishedRequest();
      }
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
        return Package.newBuilder().setName("").build();
      }
      Set<String> packages = parser.getPackages();
      if (packages.size() > 1) {
        logger.error(
            "Set of classes in {} should have only one package, instead is: {}",
            request.getRel(),
            packages);
        throw new StatusRuntimeException(Status.INVALID_ARGUMENT);
      } else if (packages.isEmpty()) {
        logger.info(
            "Set of classes in {} has no package", Paths.get(request.getRel()).toAbsolutePath());
        packages = ImmutableSet.of("");
      }
      logger.debug("Got package: {}", Iterables.getOnlyElement(packages));
      logger.debug("Got used types: {}", parser.getUsedTypes());
      logger.debug(
          "Got used packages without specific types: {}",
          parser.getUsedPackagesWithoutSpecificTypes());

      Builder packageBuilder =
          Package.newBuilder()
              .setName(Iterables.getOnlyElement(packages))
              .addAllImportedClasses(parser.getUsedTypes())
              .addAllImportedPackagesWithoutSpecificClasses(
                  parser.getUsedPackagesWithoutSpecificTypes())
              .addAllMains(parser.getMainClasses());
      for (Map.Entry<String, ImmutableSortedSet<String>> annotations :
          parser.getAnnotatedClasses().entrySet()) {
        packageBuilder.putPerClassMetadata(
            annotations.getKey(),
            PerClassMetadata.newBuilder()
                .addAllAnnotationClassNames(annotations.getValue())
                .build());
      }
      return packageBuilder.build();
    }
  }
}

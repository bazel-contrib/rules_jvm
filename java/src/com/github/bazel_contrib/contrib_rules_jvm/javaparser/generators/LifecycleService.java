package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import com.gazelle.java.javaparser.v0.LifecycleGrpc;
import com.gazelle.java.javaparser.v0.ShutdownRequest;
import com.gazelle.java.javaparser.v0.ShutdownResponse;
import io.grpc.stub.StreamObserver;

public class LifecycleService extends LifecycleGrpc.LifecycleImplBase {

  @Override
  public void shutdown(ShutdownRequest request, StreamObserver<ShutdownResponse> responseObserver) {
    // Reply to the RPC so the client doesn't block waiting for a response.
    try {
      responseObserver.onNext(ShutdownResponse.newBuilder().build());
      responseObserver.onCompleted();
    } finally {
      // Exit on a separate thread shortly after completing the response, to allow gRPC
      // to flush frames and the client to process the response without waiting on shutdown.
      new Thread(
              () -> {
                try {
                  Thread.sleep(50);
                } catch (InterruptedException ignored) {
                }
                System.exit(0);
              },
              "LifecycleShutdownExit")
          .start();
    }
  }
}

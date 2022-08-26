package com.github.bazel_contrib.contrib_rules_jvm.javaparser.generators;

import com.gazelle.java.javaparser.v0.LifecycleGrpc;
import com.gazelle.java.javaparser.v0.ShutdownRequest;
import com.gazelle.java.javaparser.v0.ShutdownResponse;
import io.grpc.stub.StreamObserver;

public class LifecycleService extends LifecycleGrpc.LifecycleImplBase {

  @Override
  public void shutdown(ShutdownRequest request, StreamObserver<ShutdownResponse> responseObserver) {
    System.exit(0);
  }
}

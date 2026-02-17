package com.example;

// NOTE: This file is in the SAME Java package as the proto-generated Options class,
// but in a DIFFERENT Bazel package.
//
// Since Options is in the same Java package, no import is needed.
// This tests that Gazelle correctly tracks:
// 1. Inner type references like Options.ResponseCode where Options is a same-package class
// 2. Proto-generated classes are indexed by their java_package

public class Client {
    // Uses Options.ResponseCode from the proto-generated class (no import needed)
    private Options.ResponseCode lastResponse;

    public Client() {
        this.lastResponse = Options.ResponseCode.SUCCESS;
    }

    public void handleResponse(Options.ResponseCode code) {
        this.lastResponse = code;
        if (code == Options.ResponseCode.INTERNAL_ERROR) {
            throw new RuntimeException("Internal error occurred");
        }
    }

    public Options.ResponseCode getLastResponse() {
        return lastResponse;
    }

    public Options.Config createConfig(String name) {
        return Options.Config.newBuilder()
            .setName(name)
            .setDefaultResponse(Options.ResponseCode.SUCCESS)
            .build();
    }
}

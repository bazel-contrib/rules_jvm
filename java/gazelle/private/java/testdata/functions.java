package com.example.test;

import org.assertj.core.api.AbstractAssert;
import org.jetbrains.annotations.NotNull;

import javax.ws.rs.core.Response;

public class ResponseAssert extends AbstractAssert<ResponseAssert, Response> {

    public ResponseAssert(Response actual) {
        super(actual, ResponseAssert.class);
    }

    public static ResponseAssert assertThat(Response actual) {
        return new ResponseAssert(actual);
    }

    public ResponseAssert isSuccessful() {
        return isStatus(200);
    }

    public ResponseAssert isCreated() {
        return isStatus(201);
    }

    public ResponseAssert isConflict() {
        return isStatus(409);
    }

    public ResponseAssert isBadRequest() {
        return isStatus(400);
    }

    public ResponseAssert isUnprocessedEntity() {
        return isStatus(422);
    }

    public ResponseAssert isUnauthorized() {
        return isStatus(403);
    }

    @NotNull
    private ResponseAssert isStatus(int expected) {
        isNotNull();

        if (this.actual.getStatus() != expected) {
            failWithMessage(
                    "Expected status to be <%s> but it was <%s>",
                    expected, this.actual.getStatus());
        }

        return this;
    }
}

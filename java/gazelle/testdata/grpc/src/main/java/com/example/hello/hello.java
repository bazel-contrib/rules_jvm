package com.example.hello;

import com.example.hello.proto.DeleteBookRequest;
import com.example.hello.proto.HelloProto;

public class Hello {
    public static void main(String[] args) {
        DeleteBookRequest req = DeleteBookRequest.newBuilder().setName("book 1").build();
        System.out.println("Hello, World!");
    }
}

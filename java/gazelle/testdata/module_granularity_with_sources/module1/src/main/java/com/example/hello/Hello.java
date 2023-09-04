package com.example.hello;

import com.example.hello.world.World;

public class Hello {
    public static void main(String[] args) {
        World one = new World();
        System.out.printf("Hello, %s!", one.name);
    }
}

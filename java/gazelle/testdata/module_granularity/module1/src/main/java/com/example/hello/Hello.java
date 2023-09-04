package com.example.hello;

import com.example.hello.world.World;
import com.example.hello.notworld.NotWorld;

public class Hello {
    public static void main(String[] args) {
        World one = new World();
        System.out.printf("Hello, %s!", one.name);
        System.out.printf("Hello also, %s!", NotWorld.NOT_WORLD);
    }
}

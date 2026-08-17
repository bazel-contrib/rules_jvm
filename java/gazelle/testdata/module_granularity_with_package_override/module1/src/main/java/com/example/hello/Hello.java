package com.example.hello;

import com.example.hello.world.World;
import com.example.hello.standalone.Standalone;

public class Hello {
    public static void main(String[] args) {
        World one = new World();
        Standalone s = new Standalone();
        System.out.printf("Hello, %s! %s", one.name, s.value);
    }
}

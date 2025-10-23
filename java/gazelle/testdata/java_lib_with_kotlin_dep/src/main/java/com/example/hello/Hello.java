package com.example.hello;

import com.example.hello.greeter.Greeter;

public class Hello {
    public static void sayHi() {
        Greeter greeter = new Greeter();
        System.out.println(greeter.greet());
    }
}

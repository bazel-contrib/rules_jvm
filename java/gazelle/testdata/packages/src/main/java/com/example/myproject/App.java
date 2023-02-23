package com.example.myproject;

import com.example.library.Library;

/** This application compares two numbers, using the Ints.compare method from Guava. */
public class App {

    public static void main(String... args) throws Exception {
        App app = new App();
        System.out.println("Success: " + Library.compare(2, 1));
    }
}

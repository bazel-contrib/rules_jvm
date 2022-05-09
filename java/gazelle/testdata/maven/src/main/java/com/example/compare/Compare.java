package com.example.compare;

import com.google.common.primitives.*;

/** This application compares two numbers, using the Ints.compare method from Guava. */
public class Compare {

    public static int compare(int a, int b) {
        return Ints.compare(a, b);
    }

    public static void main(String... args) throws Exception {
        Compare app = new Compare();
        System.out.println("Success: " + app.compare(2, 1));
    }
}

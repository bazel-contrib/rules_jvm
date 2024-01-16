package com.example.hastestutil;

public class RandomNumberGenerator {
    public static int generateNumberLessThanTwo() {
        Random random = new Random();
        return random.nextInt(2);
    }
}

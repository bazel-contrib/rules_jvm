package com.example.app;

import com.example.uninterruptibles.UninterruptibleListener;

public class App {
    public void doWork() {
        UninterruptibleListener.run(() -> System.out.println("Hello"));
    }
}

package com.example;

import java.beans.Transient;

public class Main {
    public boolean isNonNull(Transient field) {
        return field != null;
    }
}

package com.example.time;

// This class has a public static inner class that can be referenced
// by simple name from other classes in the same Java package.

public class Errors {
    private String message;

    public Errors(String message) {
        this.message = message;
    }

    public void throwIfPresent() {
        if (message != null) {
            throw new ErrorsPresentException(this);
        }
    }

    @Override
    public String toString() {
        return message;
    }

    // Public static inner class - can be accessed as just "ErrorsPresentException"
    // from other classes in the same Java package without imports.
    public static class ErrorsPresentException extends RuntimeException {
        private final Errors errors;

        public ErrorsPresentException(Errors errors) {
            super(errors.toString());
            this.errors = errors;
        }

        public Errors getErrors() {
            return errors;
        }
    }
}

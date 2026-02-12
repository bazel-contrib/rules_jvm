package com.example;

// RateLimiter should resolve via resolve_regexp to guava
import com.google.common.util.concurrent.RateLimiter;
// SpecialClass should resolve via explicit resolve directive to //special:target
import com.google.common.util.concurrent.SpecialClass;

public class Example {
    private RateLimiter limiter;
    private SpecialClass special;
}

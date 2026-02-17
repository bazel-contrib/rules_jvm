package com.example.time;

public class FakeClock implements Clock {
    private long time = 0;

    @Override
    public long currentTimeMillis() {
        return time;
    }

    public void setTime(long time) {
        this.time = time;
    }
}

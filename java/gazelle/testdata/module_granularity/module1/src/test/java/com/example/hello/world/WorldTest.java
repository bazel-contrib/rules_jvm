package com.example.hello.world;

import static org.junit.jupiter.api.Assertions.assertEquals;

import com.example.hello.notworld.justhelpersinmodule.Helper;
import org.junit.jupiter.api.Test;

public class WorldTest {

    @Test
    public void testWorld() {
        assertEquals("", "World", World.World());
    }

    @Test
    public void testNotWorld() {
        assertEquals("Not World!", Helper.getExpectation());
    }
}

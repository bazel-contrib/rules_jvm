package com.example.hello.world;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.equalTo;

import org.junit.jupiter.api.Test;

public class WorldTest {

    @Test
    public void testWorld() {
        assertThat(World.World(), equalTo("World"));
    }
}

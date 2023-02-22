package com.example.myproject;

import com.example.library.Library;

import static org.junit.Assert.assertEquals;

import org.junit.Test;

/** Tests for correct dependency retrieval with maven rules. */
public class AppTest {

    @Test
    public void testCompare() throws Exception {
        App app = new App();
        assertEquals("should return 0 when both numbers are equal", 0, Library.compare(1, 1));
    }
}

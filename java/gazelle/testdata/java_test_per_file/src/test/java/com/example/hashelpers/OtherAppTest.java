package com.example.hashelpers;

import static org.hamcrest.MatcherAssert.assertThat;
import static org.hamcrest.Matchers.equalTo;

import com.example.myproject.App;

import org.junit.Test;

public class OtherAppTest {

    @Test
    public void testCompare() throws Exception {
        App app = new App();
        assertThat(app.compare(1, 1), equalTo(0));
    }
}

package com.example.kotlinapp

import com.example.lib.Production
import com.example.lib.topLevelHelper
import org.junit.Test

class KotlinAppTest {
  @Test
  fun testPasses() {
    Production()
    topLevelHelper()
  }
}

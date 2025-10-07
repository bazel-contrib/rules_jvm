package com.example.provider

import com.google.common.base.Strings
import com.google.gson.Gson

/**
 * Custom class with componentN() functions that use external dependencies. Any code that
 * destructures this class should transitively depend on:
 * - com.google.gson (for Gson)
 * - com.google.guava (for Strings)
 */
class CustomContainer(private val data: Map<String, Any>) {

  /** component1() function that uses Gson. Destructuring should add Gson to exports. */
  operator fun component1(): String {
    val gson = Gson()
    return gson.toJson(data["first"] ?: "")
  }

  /** component2() function that uses Guava Strings. Destructuring should add Guava to exports. */
  operator fun component2(): String {
    val value = data["second"]?.toString() ?: ""
    return Strings.padEnd(value, 20, '-')
  }

  /** component3() function that uses both dependencies. */
  operator fun component3(): String {
    val gson = Gson()
    val rawValue = data["third"]?.toString() ?: ""
    val paddedValue = Strings.padStart(rawValue, 15, '*')
    return gson.toJson(mapOf("padded" to paddedValue))
  }

  /** Regular method for comparison - should not affect destructuring dependencies. */
  fun getSize(): Int = data.size
}

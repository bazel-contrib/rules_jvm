package com.example.consumer

import com.example.provider.CustomContainer

/**
 * Consumer that uses destructuring on CustomContainer. This should automatically get transitive
 * dependencies from the provider's componentN() functions through the exports mechanism.
 */
class Consumer {

  fun useDestructuring(): String {
    val container =
        CustomContainer(mapOf("first" to "hello", "second" to "world", "third" to "test"))

    // Destructure the container - this should work because
    // the provider exports its componentN() dependencies
    val (first, second, third) = container

    return "First: $first, Second: $second, Third: $third"
  }

  fun usePartialDestructuring(): String {
    val container = CustomContainer(mapOf("first" to mapOf("key" to "value"), "second" to "simple"))

    // Partial destructuring - only uses component1() and component2()
    val (jsonData, paddedData) = container

    return "JSON: $jsonData, Padded: $paddedData"
  }
}

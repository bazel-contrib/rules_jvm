package com.example.delegate

/**
 * Simple property delegates using Kotlin stdlib. These should not add any external dependencies
 * since lazy is part of Kotlin stdlib.
 */
class DelegateUtils {

  /** Lazy property delegate - should not add external dependencies. */
  val lazyValue: String by lazy { "computed value" }

  /** Another lazy property with computation. */
  val lazyNumber: Int by lazy { 42 * 2 }

  /** Regular property for comparison - should not affect dependencies. */
  val regularProperty: String = "regular value"

  fun useProperties(): String {
    return "Lazy: $lazyValue, Number: $lazyNumber, Regular: $regularProperty"
  }
}

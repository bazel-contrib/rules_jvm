package com.example.top

import com.example.middle.MiddleDelegate

/**
 * Top-level consumer that uses MiddleDelegate. This should automatically get transitive access to
 * both:
 * - commons-lang3 (from MiddleDelegate's exports)
 * - gson (from BaseDelegate's exports via MiddleDelegate's dependency)
 *
 * This demonstrates the transitive nature of the exports mechanism for property delegate
 * dependencies.
 */
class TopConsumer {

  private val middle = MiddleDelegate()

  fun processData(): String {
    // This should work because we get transitive access to all dependencies
    // through the exports mechanism
    return middle.processString("hello world")
  }
}

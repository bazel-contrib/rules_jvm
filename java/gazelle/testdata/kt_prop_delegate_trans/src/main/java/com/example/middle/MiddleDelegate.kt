package com.example.middle

import com.example.base.BaseDelegate
import com.google.common.base.Strings

/**
 * Middle layer that uses BaseDelegate and adds its own property delegates. This should export its
 * own delegate dependencies (guava) while inheriting transitive dependencies from BaseDelegate
 * (gson).
 */
class MiddleDelegate {

  private val base = BaseDelegate()

  /** Lazy property delegate that uses Guava Strings. */
  val stringProcessor: String by lazy { Strings.padEnd("processed", 20, '-') }

  fun processString(input: String): String {
    val padded = Strings.padStart(input, 15, '*')
    val json = base.toJson(mapOf("processed" to padded))
    return json
  }
}

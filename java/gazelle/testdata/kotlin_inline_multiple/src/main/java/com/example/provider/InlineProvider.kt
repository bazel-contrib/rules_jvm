package com.example.provider

import com.google.common.collect.Lists
import com.google.gson.Gson
import java.util.ArrayList

/**
 * Multiple inline functions with different dependency patterns. This tests the system's ability to
 * handle multiple inline functions in the same file with different transitive dependencies.
 */
object InlineProvider {

  /** Inline function that uses Java stdlib. */
  inline fun <T> withArrayList(block: (ArrayList<T>) -> Unit): ArrayList<T> {
    val list = ArrayList<T>()
    block(list)
    return list
  }

  /** Inline function that uses Gson for JSON processing. */
  inline fun toJson(data: Any): String {
    val gson = Gson()
    return gson.toJson(data)
  }

  /** Inline function that uses Guava collections. */
  inline fun <T> createGuavaList(vararg items: T): List<T> {
    return Lists.newArrayList(*items)
  }

  /** Inline function that combines multiple dependencies. */
  inline fun processAndSerialize(data: List<String>): String {
    val arrayList = ArrayList(data)
    val guavaList = Lists.newArrayList(arrayList)
    val gson = Gson()
    return gson.toJson(guavaList)
  }

  /** Simple inline function with no external dependencies. */
  inline fun calculate(x: Int, y: Int): Int {
    return x * y + 42
  }

  /** Regular function - should not affect transitive dependencies. */
  fun regularFunction(input: String): String {
    val gson = Gson()
    return gson.toJson(input)
  }
}

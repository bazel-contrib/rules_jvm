package com.example.multiple

import com.google.common.base.Strings
import com.google.gson.Gson
import java.util.concurrent.ConcurrentHashMap
import kotlin.properties.Delegates

/**
 * Class with multiple property delegates using different external dependencies. This should export
 * all the dependencies used by the delegates:
 * - com.google.gson (for Gson)
 * - com.google.guava (for Strings)
 * - java.util.concurrent (for ConcurrentHashMap)
 */
class MultipleDelegates {

  /** Lazy delegate using Gson. */
  val jsonProcessor: Gson by lazy { Gson() }

  /** Lazy delegate using Guava Strings utility. */
  val stringProcessor: String by lazy { Strings.padEnd("hello", 10, ' ') }

  /** Observable delegate using ConcurrentHashMap. */
  var dataMap: Map<String, String> by
      Delegates.observable(emptyMap<String, String>()) { _, _, newValue ->
        val concurrentMap = ConcurrentHashMap<String, String>()
        concurrentMap.putAll(newValue)
        println("Map updated with ${concurrentMap.size} entries")
      }

  /** Vetoable delegate using multiple dependencies. */
  var validatedJson: String by
      Delegates.vetoable("{}") { _, _, newValue ->
        // Use Guava Strings to check if not null or empty
        if (Strings.isNullOrEmpty(newValue)) {
          false
        } else {
          // Use Gson to validate JSON
          try {
            val gson = Gson()
            gson.fromJson(newValue, Map::class.java)
            true
          } catch (e: Exception) {
            false
          }
        }
      }

  /** Another lazy delegate with complex initialization. */
  val complexProcessor: String by lazy {
    val gson = Gson()
    val processed = Strings.padStart("hello world", 20, '*')
    gson.toJson(mapOf("processed" to processed))
  }

  fun processAll(): String {
    val json = jsonProcessor.toJson("test")
    val padded = Strings.padEnd("test", 10, '-')
    dataMap = mapOf("key1" to "value1", "key2" to "value2")
    validatedJson = "{\"test\": \"value\"}"

    return "JSON: $json, Padded: $padded, Complex: $complexProcessor"
  }
}

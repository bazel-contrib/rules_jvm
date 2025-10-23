package com.example.consumer

import com.example.provider.InlineProvider.calculate
import com.example.provider.InlineProvider.createGuavaList
import com.example.provider.InlineProvider.processAndSerialize
import com.example.provider.InlineProvider.regularFunction
import com.example.provider.InlineProvider.toJson
import com.example.provider.InlineProvider.withArrayList

/**
 * Consumer that uses multiple inline functions from the provider. This should get transitive
 * dependencies from all the inline functions it uses.
 */
class Consumer {

  fun useArrayListInline(): String {
    val result =
        withArrayList<String> { list ->
          list.add("test")
          list.add("data")
        }
    return result.toString()
  }

  fun useJsonInline(): String {
    val data = mapOf("key" to "value")
    return toJson(data)
  }

  fun useGuavaInline(): List<String> {
    return createGuavaList("a", "b", "c")
  }

  fun useComplexInline(): String {
    val data = listOf("item1", "item2", "item3")
    return processAndSerialize(data)
  }

  fun useSimpleInline(): Int {
    return calculate(5, 10)
  }

  fun useRegularFunction(): String {
    return regularFunction("test")
  }
}

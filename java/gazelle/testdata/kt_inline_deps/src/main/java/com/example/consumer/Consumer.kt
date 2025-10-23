package com.example.consumer

import com.example.provider.processData
import com.example.provider.regularProcessData

/**
 * Consumer that calls both inline and regular functions. This should:
 * - Have transitive dependencies from processData (inline function)
 * - NOT have transitive dependencies from regularProcessData (regular function)
 */
class Consumer {
  fun useInlineFunction(): String {
    return processData("test data")
  }

  fun useRegularFunction(): String {
    return regularProcessData("test data")
  }
}

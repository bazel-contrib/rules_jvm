package com.example.level2

import com.example.level1.Level1.createList
import com.example.level1.Level1.multiply
import com.google.gson.Gson

/**
 * Level 2 inline functions - these call other inline functions AND use their own dependencies.
 * This tests transitive inline function calls.
 */
object Level2 {
    
    /**
     * Inline function that calls another inline function (createList) AND uses Gson.
     * This should cause transitive dependencies from both:
     * 1. Its own Gson usage
     * 2. The ArrayList dependency from createList
     */
    inline fun createAndSerialize(vararg items: String): String {
        val list = createList(*items)  // Calls inline function from Level1
        val gson = Gson()               // Uses its own dependency
        return gson.toJson(list)
    }

    /**
     * Inline function that calls multiple other inline functions.
     */
    inline fun processNumbers(x: Int, y: Int): String {
        val result = multiply(x, y)     // Calls inline function from Level1
        val list = createList(result)   // Calls another inline function from Level1
        val gson = Gson()               // Uses its own dependency
        return gson.toJson(list)
    }

    /**
     * Regular function for comparison.
     */
    fun regularProcess(data: String): String {
        val gson = Gson()
        return gson.toJson(data)
    }
}

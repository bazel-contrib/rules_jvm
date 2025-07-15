package com.example.consumer

import com.example.extension.parseToJsonObject
import com.example.extension.processJsonData
import com.example.extension.reverseString

/**
 * Consumer that uses extension functions from another package.
 * This should get transitive dependencies from the extension functions.
 */
class Consumer {
    fun useExtensions() {
        // This should pull in Gson and JsonObject dependencies from the extension function
        val jsonObj = "{"test": "value"}".parseToJsonObject()
        println("Parsed: ${jsonObj}")
        
        // This should pull in multiple Gson class dependencies
        val processed = "hello".processJsonData()
        println("Processed: $processed")
        
        // This should not add any extra dependencies
        val reversed = "world".reverseString()
        println("Reversed: $reversed")
    }
}

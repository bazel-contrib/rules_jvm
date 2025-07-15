package com.example.base

import com.google.gson.Gson

/**
 * Base class with property delegates that use external dependencies.
 * This exports Gson so that consumers can use it transitively.
 */
class BaseDelegate {
    
    /**
     * Lazy property delegate that uses Gson.
     */
    val jsonProcessor: Gson by lazy {
        Gson()
    }
    
    fun toJson(data: Any): String {
        return jsonProcessor.toJson(data)
    }
}

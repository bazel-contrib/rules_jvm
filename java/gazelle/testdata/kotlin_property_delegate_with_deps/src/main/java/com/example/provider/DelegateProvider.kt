package com.example.provider

import com.google.gson.Gson
import kotlin.properties.Delegates
import java.util.ArrayList

/**
 * Property delegates that use external dependencies.
 * Any code that uses these properties should transitively depend on:
 * - com.google.gson (for Gson)
 * - java.util (for ArrayList)
 */
class DelegateProvider {
    
    /**
     * Lazy property delegate that uses external dependencies.
     * The lazy block uses Gson, so consumers should get Gson transitively.
     */
    val jsonProcessor: Gson by lazy {
        Gson()
    }
    
    /**
     * Observable property delegate that uses external dependencies.
     * The observer uses ArrayList, so consumers should get java.util transitively.
     */
    var dataList: List<String> by Delegates.observable(emptyList<String>()) { _, oldValue, newValue ->
        val list = ArrayList<String>()
        list.addAll(oldValue)
        list.addAll(newValue)
        println("Data changed from ${list.size - newValue.size} to ${list.size} items")
    }
    
    /**
     * Vetoable property delegate with external dependencies.
     */
    var validatedData: String by Delegates.vetoable("") { _, _, newValue ->
        val gson = Gson()
        // Use Gson to validate the new value
        try {
            gson.fromJson(newValue, String::class.java)
            true
        } catch (e: Exception) {
            false
        }
    }
    
    /**
     * Regular property for comparison - should not affect dependencies.
     */
    val regularProperty: String = "regular value"
    
    fun processData(input: String): String {
        return jsonProcessor.toJson(input)
    }
}

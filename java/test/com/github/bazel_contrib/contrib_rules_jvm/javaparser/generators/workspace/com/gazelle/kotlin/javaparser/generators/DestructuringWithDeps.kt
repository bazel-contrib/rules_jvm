package workspace.com.gazelle.kotlin.javaparser.generators

import com.google.gson.Gson
import com.google.common.base.Strings

/**
 * Test file for destructuring declarations with custom componentN() functions
 * that have external dependencies.
 */

/**
 * Custom class with componentN() functions that use external dependencies.
 * Any code that destructures this class should transitively depend on:
 * - com.google.gson (for Gson)
 * - com.google.guava (for Strings)
 */
class CustomContainer(private val data: Map<String, Any>) {
    
    /**
     * component1() function that uses Gson.
     * Destructuring should add Gson to exports.
     */
    operator fun component1(): String {
        val gson = Gson()
        return gson.toJson(data["first"] ?: "")
    }
    
    /**
     * component2() function that uses Guava Strings.
     * Destructuring should add Guava to exports.
     */
    operator fun component2(): String {
        val value = data["second"]?.toString() ?: ""
        return Strings.padEnd(value, 20, '-')
    }
    
    /**
     * component3() function that uses both dependencies.
     */
    operator fun component3(): String {
        val gson = Gson()
        val rawValue = data["third"]?.toString() ?: ""
        val paddedValue = Strings.padStart(rawValue, 15, '*')
        return gson.toJson(mapOf("padded" to paddedValue))
    }
    
    /**
     * Regular method for comparison - should not affect destructuring dependencies.
     */
    fun getSize(): Int = data.size
}

/**
 * Consumer that uses destructuring on CustomContainer.
 * This should automatically get transitive dependencies from the provider's componentN() functions
 * through the exports mechanism.
 */
class Consumer {
    
    fun useDestructuring(): String {
        val container = CustomContainer(mapOf(
            "first" to "hello",
            "second" to "world",
            "third" to "test"
        ))
        
        // Destructure the container - this should work because
        // the provider exports its componentN() dependencies
        val (first, second, third) = container
        
        return "First: $first, Second: $second, Third: $third"
    }
    
    fun usePartialDestructuring(): String {
        val container = CustomContainer(mapOf(
            "first" to mapOf("key" to "value"),
            "second" to "simple"
        ))
        
        // Partial destructuring - only uses component1() and component2()
        val (jsonData, paddedData) = container
        
        return "JSON: $jsonData, Padded: $paddedData"
    }
}

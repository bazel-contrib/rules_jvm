package workspace.com.gazelle.kotlin.javaparser.generators

import com.example.Helper
import com.google.gson.Gson
import com.google.gson.JsonArray
import com.google.gson.JsonElement

/**
 * Test file with simple extension functions that use external dependencies.
 */

// Extension function on String that uses Gson
fun String.parseAsJson(): com.google.gson.JsonObject {
    val gson = Gson()
    return gson.fromJson(this, com.google.gson.JsonObject::class.java)
}

// Extension function on String that uses JsonArray (another Gson class)
fun String.parseAsJsonArray(): JsonArray {
    val gson = Gson()
    return gson.fromJson(this, JsonArray::class.java)
}

// Extension function on String that uses Helper
fun String.processWithHelper(): String {
    val helper = Helper()
    return helper.transform(this)
}

// Extension function with no external dependencies
fun String.reverseString(): String {
    return this.reversed()
}

// Regular function for comparison (not an extension)
fun regularFunction(data: String): String {
    return data.uppercase()
}

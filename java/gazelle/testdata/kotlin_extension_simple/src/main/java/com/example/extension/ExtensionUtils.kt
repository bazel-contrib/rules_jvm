package com.example.extension

import com.google.gson.Gson
import com.google.gson.JsonObject

/**
 * Simple extension function that uses Gson library.
 * This should cause any code that calls this function to depend on Gson classes.
 */
fun String.parseToJsonObject(): JsonObject {
    val gson = Gson()
    return gson.fromJson(this, JsonObject::class.java)
}

/**
 * Extension function that uses multiple Gson classes.
 */
fun String.processJsonData(): String {
    val gson = Gson()
    val jsonObj = JsonObject()
    jsonObj.addProperty("input", this)
    return gson.toJson(jsonObj)
}

/**
 * Simple extension function with no external dependencies.
 * This should not add any extra dependencies to calling code.
 */
fun String.reverseString(): String {
    return this.reversed()
}

/**
 * Regular function for comparison - should not affect dependencies.
 */
fun regularFunction(data: String): String {
    return data.uppercase()
}

package com.example.provider

import java.util.ArrayList
import com.google.gson.Gson

/**
 * Inline function that uses external dependencies.
 * Any code that calls this function should transitively depend on:
 * - java.util (for ArrayList)
 * - com.google.gson (for Gson)
 */
inline fun processData(data: String): String {
    val list = ArrayList<String>()
    list.add(data)
    
    val gson = Gson()
    return gson.toJson(list)
}

/**
 * Regular function that also uses external dependencies.
 * This should NOT cause transitive dependencies for callers.
 */
fun regularProcessData(data: String): String {
    val gson = Gson()
    return gson.toJson(data)
}

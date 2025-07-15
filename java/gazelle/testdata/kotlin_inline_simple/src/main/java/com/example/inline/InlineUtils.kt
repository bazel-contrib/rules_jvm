package com.example.inline

import java.util.ArrayList

/**
 * Simple inline function that uses Java stdlib.
 * This should cause any code that calls this function to depend on java.util.
 */
inline fun <T> withList(block: (ArrayList<T>) -> Unit) {
    val list = ArrayList<T>()
    block(list)
}

/**
 * Simple inline function with no external dependencies.
 * This should not add any extra dependencies to calling code.
 */
inline fun simpleAdd(x: Int, y: Int): Int {
    return x + y
}

/**
 * Regular function for comparison - should not affect dependencies.
 */
fun regularFunction(data: String): String {
    return data.uppercase()
}

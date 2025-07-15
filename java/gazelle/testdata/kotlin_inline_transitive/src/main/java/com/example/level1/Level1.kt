package com.example.level1

import java.util.ArrayList

/**
 * Level 1 inline functions - these use basic dependencies.
 */
object Level1 {
    
    /**
     * Basic inline function that uses ArrayList.
     */
    inline fun <T> createList(vararg items: T): ArrayList<T> {
        val list = ArrayList<T>()
        for (item in items) {
            list.add(item)
        }
        return list
    }

    /**
     * Simple inline function with no external dependencies.
     */
    inline fun multiply(x: Int, y: Int): Int {
        return x * y
    }
}

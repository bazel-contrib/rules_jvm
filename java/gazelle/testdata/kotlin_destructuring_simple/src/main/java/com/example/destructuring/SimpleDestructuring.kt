package com.example.destructuring

/**
 * Simple destructuring using data classes.
 * Data classes automatically generate componentN() functions,
 * so this should not add any external dependencies.
 */
data class Point(val x: Int, val y: Int)

data class Person(val name: String, val age: Int, val email: String)

class SimpleDestructuring {
    
    fun useDestructuring(): String {
        val point = Point(10, 20)
        val (x, y) = point  // Uses component1() and component2()
        
        val person = Person("Alice", 30, "alice@example.com")
        val (name, age, email) = person  // Uses component1(), component2(), component3()
        
        // Partial destructuring
        val (personName, _) = person  // Only uses component1()
        
        return "Point: ($x, $y), Person: $name, $age, $email, Name only: $personName"
    }
    
    fun useInLoop(): List<String> {
        val points = listOf(Point(1, 2), Point(3, 4), Point(5, 6))
        
        return points.map { (x, y) ->  // Destructuring in lambda
            "($x, $y)"
        }
    }
}

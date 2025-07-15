package workspace.com.gazelle.kotlin.javaparser.generators

import com.example.MathUtils
import com.google.gson.JsonObject
import com.google.gson.JsonArray

/**
 * Test file with extension operators that use external dependencies.
 */

data class Matrix(val data: Array<Array<Double>>)

// Extension operator that uses MathUtils
operator fun Matrix.plus(other: Matrix): Matrix {
    val mathUtils = MathUtils()
    return mathUtils.addMatrices(this, other)
}

// Extension operator that uses Gson classes
operator fun Matrix.times(scalar: Double): Matrix {
    val jsonObj = JsonObject()
    val jsonArray = JsonArray()
    // Use Gson classes in the computation
    jsonObj.addProperty("scalar", scalar)
    jsonArray.add(jsonObj)
    
    // Process with Gson classes
    return Matrix(this.data.map { row -> row.map { it * scalar }.toTypedArray() }.toTypedArray())
}

// Extension operator with no external dependencies
operator fun Matrix.unaryMinus(): Matrix {
    return Matrix(this.data.map { row -> row.map { -it }.toTypedArray() }.toTypedArray())
}

// Regular function for comparison (not an operator extension)
fun multiplyMatrices(a: Matrix, b: Matrix): Matrix {
    return Matrix(emptyArray())
}

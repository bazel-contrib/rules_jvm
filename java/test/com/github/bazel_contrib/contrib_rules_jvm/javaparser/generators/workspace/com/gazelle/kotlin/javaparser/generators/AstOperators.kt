package workspace.com.gazelle.kotlin.javaparser.generators

import com.example.MathUtils
import com.google.gson.JsonObject

class Vector(val x: Int, val y: Int)

// Arithmetic operators
operator fun Vector.plus(other: Vector): Vector {
  val utils = MathUtils()
  return utils.addVectors(this, other)
}

operator fun Vector.minus(other: Vector): Vector {
  return Vector(this.x - other.x, this.y - other.y)
}

operator fun Vector.times(scalar: Int): Vector {
  val json = JsonObject()
  json.addProperty("operation", "scale")
  return Vector(this.x * scalar, this.y * scalar)
}

// Unary operators
operator fun Vector.unaryMinus(): Vector = Vector(-x, -y)

operator fun Vector.not(): Boolean = x == 0 && y == 0

// Comparison operators
operator fun Vector.compareTo(other: Vector): Int {
  val thisLength = x * x + y * y
  val otherLength = other.x * other.x + other.y * other.y
  return thisLength.compareTo(otherLength)
}

// Usage of operators
fun testOperators() {
  val v1 = Vector(1, 2)
  val v2 = Vector(3, 4)

  val sum = v1 + v2 // plus operator
  val diff = v1 - v2 // minus operator
  val scaled = v1 * 3 // times operator
  val negated = -v1 // unaryMinus operator
  val isEmpty = !v1 // not operator
  val comparison = v1 > v2 // compareTo operator
}

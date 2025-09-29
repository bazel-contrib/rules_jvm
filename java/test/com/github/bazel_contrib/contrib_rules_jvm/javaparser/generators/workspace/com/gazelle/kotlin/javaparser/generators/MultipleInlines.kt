package workspace.com.gazelle.kotlin.javaparser.generators

import com.example.utils.StringUtils
import java.util.ArrayList

/** Test file with multiple inline functions to verify we can detect all of them. */
object MultipleInlines {

  // Inline function using Java stdlib
  inline fun <T> withList(block: (ArrayList<T>) -> Unit) {
    val list = ArrayList<T>()
    block(list)
  }

  // Inline function using external dependency
  inline fun formatString(input: String): String {
    val utils = StringUtils()
    return utils.format(input)
  }

  // Inline function with no external dependencies
  inline fun simpleInline(x: Int, y: Int): Int {
    return x + y
  }

  // Regular function (should not be detected as inline)
  fun notInline(data: String): String {
    return data.lowercase()
  }
}

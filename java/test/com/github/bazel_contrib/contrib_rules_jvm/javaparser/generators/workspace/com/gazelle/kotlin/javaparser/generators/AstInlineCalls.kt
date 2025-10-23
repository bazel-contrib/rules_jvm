package workspace.com.gazelle.kotlin.javaparser.generators

import com.example.Helper
import com.google.gson.Gson

// Define inline functions
inline fun processJson(data: String): String {
  val gson = Gson()
  return gson.toJson(data)
}

inline fun processHelper(input: String): String {
  val helper = Helper()
  return helper.process(input)
}

inline fun unusedInlineFunction(x: Int): Int = x * 2

// Functions that use inline functions
fun actualUsage() {
  val result1 = processJson("test") // This should be detected
  val result2 = processHelper("data") // This should be detected
  // Note: unusedInlineFunction is NOT called
}

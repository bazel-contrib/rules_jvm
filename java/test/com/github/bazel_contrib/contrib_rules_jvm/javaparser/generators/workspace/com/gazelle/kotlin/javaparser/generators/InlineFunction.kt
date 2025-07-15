package workspace.com.gazelle.kotlin.javaparser.generators

import com.example.Helper
import com.google.gson.Gson

/**
 * Test file with a simple inline function that uses external dependencies.
 */
class InlineFunctionExample {
    
    inline fun processData(data: String): String {
        val helper = Helper()
        val gson = Gson()
        return gson.toJson(helper.transform(data))
    }
    
    // Regular function for comparison
    fun regularFunction(data: String): String {
        return data.uppercase()
    }
}

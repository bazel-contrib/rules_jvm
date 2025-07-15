package workspace.com.gazelle.kotlin.javaparser.generators;

import com.google.gson.Gson;
import com.example.Helper;

// Inline function
inline fun processData(data: String): String {
    val gson = Gson()
    return gson.toJson(data)
}

// Extension function
fun String.transform(): String {
    val helper = Helper()
    return helper.process(this)
}

// Complex usage patterns
fun testComplexPatterns() {
    val data = "test"
    
    // Nested calls
    val result1 = processData(data.transform())
    
    // Chained calls
    val result2 = data.transform().transform()
    
    // Mixed inline and extension
    val result3 = processData(data).transform()
    
    // Safe chained calls
    val nullable: String? = "test"
    val result4 = nullable?.transform()?.transform()
}

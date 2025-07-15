package workspace.com.gazelle.kotlin.javaparser.generators;

import com.google.gson.Gson;
import com.example.StringProcessor;

// Extension functions on different types
fun String.processAsString(): String {
    val gson = Gson()
    return gson.toJson(this)
}

fun Int.processAsInt(): String {
    val processor = StringProcessor()
    return processor.format(this.toString())
}

fun List<String>.processAsList(): String {
    return this.joinToString(",")
}

// Usage with different receiver types
fun testReceivers() {
    val str = "hello"
    val num = 42
    val list = listOf("a", "b", "c")
    
    val result1 = str.processAsString()    // String receiver
    val result2 = num.processAsInt()       // Int receiver  
    val result3 = list.processAsList()     // List receiver
}

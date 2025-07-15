package workspace.com.gazelle.kotlin.javaparser.generators;

import com.google.gson.Gson;

// Extension functions on nullable types
fun String?.safeProcess(): String? {
    val gson = Gson()
    return this?.let { gson.toJson(it) }
}

fun String?.safeLength(): Int? {
    return this?.length
}

// Usage with safe calls
fun testSafeCalls() {
    val nullable: String? = "test"
    val alsoNullable: String? = null
    
    val result1 = nullable?.safeProcess()      // Safe call on extension
    val result2 = alsoNullable?.safeProcess()  // Safe call on null
    val length1 = nullable?.safeLength()       // Another safe call
    val length2 = alsoNullable?.safeLength()   // Safe call on null
}

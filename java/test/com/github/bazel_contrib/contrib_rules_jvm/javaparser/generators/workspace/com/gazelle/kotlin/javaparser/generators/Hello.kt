package workspace.com.gazelle.kotlin.javaparser.generators

import com.gazelle.java.javaparser.generators.DeleteBookRequest
import com.gazelle.java.javaparser.generators.HelloProto
import com.google.common.primitives.Ints

class Hello {
  fun useImports() {
    val req: DeleteBookRequest? = null
    val proto: HelloProto? = null
    val bytes = Ints.BYTES
    // Touch the values so tools consider them used
    if (bytes >= 0) {
      println("$req $proto $bytes")
    }
  }
}

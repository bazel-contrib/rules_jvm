package workspace.com.gazelle.kotlin.javaparser.generators

import com.gazelle.java.javaparser.ClasspathParser.MAIN_CLASS
import com.gazelle.kotlin.constantpackage.CONSTANT
import com.gazelle.kotlin.constantpackage2.FOO
import com.gazelle.kotlin.functionpackage.someFunction

fun myFunction() {
  // Use the imported constants so ktfmt or optimize imports doesn't drop them
  val s = "${CONSTANT}_${FOO}_$MAIN_CLASS"
  println(s)
  someFunction()
}

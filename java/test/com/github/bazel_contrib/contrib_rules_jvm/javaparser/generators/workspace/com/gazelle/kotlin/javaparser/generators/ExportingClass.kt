package workspace.com.gazelle.kotlin.javaparser.generators

import example.external.ParameterizedReturn
import example.external.InternalReturn
import example.external.PrivateReturn
import example.external.ProtectedReturn
import example.external.PublicReturn

class ExportingClass {
  internal fun getInternal(): InternalReturn? {
    return null
  }

  private fun getPrivate(): PrivateReturn? {
    return null
  }

  protected fun getProtected(): ProtectedReturn? {
    return null
  }

  fun getPublic(): PublicReturn? {
    return null
  }

  fun getPrimitive(): Int {
    return 0
  }

  fun getVoid(): Unit {
  }

  fun getParameterizedType(): ParameterizedReturn<String>? {
    return null
  }
}


package workspace.com.gazelle.kotlin.javaparser.generators

import example.external.FinalProperty
import example.external.VarProperty
import example.external.InternalReturn
import example.external.ProtectedReturn
import example.external.PublicReturn

private class PrivateExportingClass {
  val finalProperty: FinalProperty? = null
  val varProperty: VarProperty? = null

  internal fun getInternal(): InternalReturn? {
    return null
  }

  protected fun getProtected(): ProtectedReturn? {
    return null
  }

  fun getPublic(): PublicReturn? {
    return null
  }
}


package workspace.com.gazelle.kotlin.javaparser.generators

import example.external.ConstructorParam
import example.external.DataClassProperty
import example.external.FunctionTypeParam
import example.external.FunctionTypeReturn
import example.external.GenericArg
import example.external.GenericOuter
import example.external.InternalParam
import example.external.NullableParam
import example.external.PrivateConstructorParam
import example.external.PrivateParam
import example.external.ProtectedParam
import example.external.PublicParam
import example.external.SecondaryConstructorParam
import example.external.VarargParam

class ParameterExportingClass(position: ConstructorParam) {

  constructor(alt: SecondaryConstructorParam) : this(ConstructorParam())

  fun acceptsPublic(param: PublicParam) {}

  internal fun acceptsInternal(param: InternalParam) {}

  protected fun acceptsProtected(param: ProtectedParam) {}

  private fun acceptsPrivate(param: PrivateParam) {}

  fun acceptsVarargs(vararg params: VarargParam) {}

  fun acceptsNullable(param: NullableParam?) {}

  fun acceptsParameterized(param: GenericOuter<GenericArg>) {}

  fun acceptsFunctionType(callback: (FunctionTypeParam) -> FunctionTypeReturn) {}
}

data class ParameterDataClass(val prop: DataClassProperty)

class PrivateConstructorClass private constructor(param: PrivateConstructorParam)

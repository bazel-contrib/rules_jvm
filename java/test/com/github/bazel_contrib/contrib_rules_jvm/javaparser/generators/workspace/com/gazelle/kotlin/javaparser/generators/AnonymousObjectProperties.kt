package workspace.com.gazelle.kotlin.javaparser.generators

object PropertyContainer {
  fun build() = object {
    val property: com.example.PropertyType? = null
    var mutableProperty: com.example.MutablePropertyType? = null
    val delegatedProperty: com.example.DelegatedPropertyType by lazy {
      com.example.DelegatedPropertyType()
    }
  }
}

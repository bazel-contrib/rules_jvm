package com.example.consumer

import com.example.provider.DelegateProvider

/**
 * Consumer that uses property delegates from another package.
 * This should automatically get transitive dependencies from the provider's delegates
 * through the exports mechanism.
 */
class Consumer {
    
    private val provider = DelegateProvider()
    
    fun useDelegate(): String {
        // Access the delegated properties - this should work because
        // the provider exports its delegate dependencies
        val processed = provider.processData("test data")
        
        // Modify observable property
        provider.dataList = listOf("item1", "item2")
        
        // Try to set vetoable property
        provider.validatedData = "\"valid json string\""
        
        return "Processed: $processed, Data: ${provider.dataList}, Validated: ${provider.validatedData}"
    }
}

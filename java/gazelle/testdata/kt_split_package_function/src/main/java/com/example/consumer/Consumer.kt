package com.example.consumer

import com.example.split.getLogger

// getLogger is a top-level function in com.example.split, which is a split Maven package (provided
// by both lib-a and lib-b). A function import has no class-name segment, so package-level
// resolution alone is ambiguous; it must be resolved to lib-a at class granularity via the class
// index, which lists top-level functions under their package.
class Consumer {
  private val logger = getLogger()
}

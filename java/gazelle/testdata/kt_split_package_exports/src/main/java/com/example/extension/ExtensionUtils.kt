package com.example.extension

import com.example.split.ClassA

// ClassA comes from a split Maven package (com.example.split is provided by both lib-a and lib-b).
// Because it is the return type of this public extension function, callers depend on it transitively,
// so it must appear in both exports and deps. The split package can only be resolved to a single
// artifact at class granularity.
fun String.toClassA(): ClassA {
  TODO("$this")
}

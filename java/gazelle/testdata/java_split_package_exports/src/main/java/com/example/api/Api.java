package com.example.api;

import com.example.split.ClassA;

// ClassA comes from a split Maven package (com.example.split is provided by both lib-a and lib-b).
// Because it appears in this interface's public API, it must be both a dep and an export, and can
// only be resolved to a single artifact at class granularity.
public interface Api {
  ClassA get();
}

package com.example.foo

class Foo {
  // An internal member that a same-module (associated) test may read.
  internal fun secret(): Int = 42
}

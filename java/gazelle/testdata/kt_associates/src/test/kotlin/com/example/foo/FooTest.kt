package com.example.foo

// Same package as Foo: with associates it shares Foo's Kotlin module and can read `internal`.
class FooTest {
  fun checksSecret() {
    check(Foo().secret() == 42)
  }
}

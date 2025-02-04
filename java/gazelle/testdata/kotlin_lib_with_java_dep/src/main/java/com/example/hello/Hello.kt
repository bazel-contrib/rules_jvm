package com.example.hello

import com.example.hello.greeter.Greeter

fun sayHi() {
    val greeter = Greeter()
    println("${greeter.greet()}")
}

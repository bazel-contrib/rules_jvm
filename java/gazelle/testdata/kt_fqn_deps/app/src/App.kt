package com.example.app

fun doWork() {
    try {
        // work
    } catch (e: Exception) {
        throw com.example.errors.CustomError(e)
    }
}

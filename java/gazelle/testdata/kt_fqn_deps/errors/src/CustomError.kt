package com.example.errors

class CustomError(cause: Exception) : RuntimeException(cause)

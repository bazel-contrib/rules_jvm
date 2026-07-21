package com.example.extensions

import com.example.model.Model
import com.google.gson.JsonObject

fun String.toModel(): Model = Model()

fun String.toJson(): JsonObject = TODO()

package com.example.user

import com.example.base.Base

// Acyclic public dependency: user stays its own target with a dep on base.
class User {
  private val base = Base()
}

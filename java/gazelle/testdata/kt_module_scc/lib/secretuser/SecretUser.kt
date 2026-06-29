package com.example.secretuser

import com.example.secret.Secret

// Acyclic reference to secret's `internal` symbol: forces both into one module/target.
class SecretUser {
  private val secret = Secret()
}

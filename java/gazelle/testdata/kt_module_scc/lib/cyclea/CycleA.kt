package com.example.cyclea

import com.example.cycleb.CycleB

// Import cycle with cycleb: the two packages collapse into one target.
class CycleA {
  fun b(): CycleB? = null
}

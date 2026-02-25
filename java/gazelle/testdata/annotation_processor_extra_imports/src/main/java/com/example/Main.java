package com.example;

import com.google.auto.value.AutoValue;
import com.google.auto.value.extension.memoized.Memoized;

@AutoValue
abstract class Main {
  static Main create(String name, int numberOfLegs) {
    return new AutoValue_Main(name, numberOfLegs);
  }

  abstract String name();
  abstract int numberOfLegs();

  @Memoized
  String description() {
    return name() + " has " + numberOfLegs() + " legs";
  }
}

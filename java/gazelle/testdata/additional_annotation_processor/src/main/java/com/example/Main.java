package com.example;

import com.google.auto.value.AutoValue;

class Main {
  public static void main(String[] args) {
    Animal pig = Animal.create("pig", 4);
    Animal chicken = Animal.create("chicken", 2);
    System.out.printf("Checking if %s has same legs as %s: %s%n", pig, chicken, pig.numberOfLegs() == chicken.numberOfLegs());
  }

  @AutoValue
  public abstract static class Animal {
    static Animal create(String name, int numberOfLegs) {
      return new AutoValue_Main_Animal(name, numberOfLegs);
    }

    abstract String name();
    abstract int numberOfLegs();
  }
}

package com.example.package2;

import com.example.package1.ClassWithInnerClasses;

public class Caller {
    public void doCall() {
        ClassWithInnerClass.InnerClass innerClass = new ClassWithInnerClass.InnerClass();
        System.out.println(innerClass.getMessage());
    }
}

package com.example.consumer;

import com.example.iface.Interface;

public class Consumer {
  public void useInterface(Interface i) {
    var t = i.get();
    System.out.println(t.stringify() + "hello");
  }
}

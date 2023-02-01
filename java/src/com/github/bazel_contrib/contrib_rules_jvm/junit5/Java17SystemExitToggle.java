package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.lang.invoke.MethodHandles;
import java.util.Map;
import sun.misc.Unsafe;

public class Java17SystemExitToggle extends Java11SystemExitToggle {

  public Java17SystemExitToggle() throws ReflectiveOperationException {
    suppressSecurityManagerWarning();
  }

  private void suppressSecurityManagerWarning() throws ReflectiveOperationException {
    Class<?> holderClazz = Class.forName("java.lang.System$CallersHolder");
    var callersField = holderClazz.getDeclaredField("callers");

    var unsafe = getUnsafe();

    // And we can't just grab the field easily
    var base = unsafe.staticFieldBase(callersField);
    long offset = unsafe.staticFieldOffset(callersField);
    Object callers = unsafe.getObject(base, offset);

    // And now we inject ourselves into the SecurityManager
    if (Map.class.isAssignableFrom(callers.getClass())) {
      @SuppressWarnings("unchecked")
      Map<Class<?>, Boolean> map = Map.class.cast(callers);
      map.put(getClass().getSuperclass(), true);
    }
  }

  @Override
  public void prevent() {
    try {
      super.prevent();
    } catch (UnsupportedOperationException ignored) {
      // Thrown when the security manager has been marked as unavailable
      System.err.println(
          "Unable to disable security manager. Calls to `System.exit`"
              + " may succeed in tests. Some versions of java allow this behaviour to "
              + "be overridden by setting the `java.security.manager` to `allow`");
    }
  }

  @Override
  public void allow() {
    try {
      super.allow();
    } catch (UnsupportedOperationException ignored) {
      // Thrown when the security manager has been marked as unavailable
      // We have already warned the user about this in the call to `prevent`
      // so silently fall through
    }
  }

  private Unsafe getUnsafe() throws ReflectiveOperationException {
    // We want to avoid opening java.lang, so let's jump through hoops

    // First problem: we can't call `Unsafe.getUnsafe` directly.
    var c = Class.forName("sun.misc.Unsafe");
    var lookup = MethodHandles.privateLookupIn(c, MethodHandles.lookup());
    var handle = lookup.findStaticVarHandle(c, "theUnsafe", c);
    return (Unsafe) handle.get();
  }
}

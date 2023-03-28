package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.lang.reflect.Field;
import java.util.Map;
import sun.misc.Unsafe;

public class Java17SystemExitToggle implements SystemExitToggle {

  private final SystemExitToggle toggle;

  public Java17SystemExitToggle(SystemExitToggle toggle) throws ReflectiveOperationException {
    this.toggle = toggle;
    suppressSecurityManagerWarning();
  }

  private void suppressSecurityManagerWarning() throws ReflectiveOperationException {
    Class<?> holderClazz = Class.forName("java.lang.System$CallersHolder");
    Field callersField = holderClazz.getDeclaredField("callers");

    Unsafe unsafe = getUnsafe();

    // And we can't just grab the field easily
    Object base = unsafe.staticFieldBase(callersField);
    long offset = unsafe.staticFieldOffset(callersField);
    Object callers = unsafe.getObject(base, offset);

    // And now we inject ourselves into the SecurityManager
    if (Map.class.isAssignableFrom(callers.getClass())) {
      @SuppressWarnings("unchecked")
      Map<Class<?>, Boolean> map = Map.class.cast(callers);
      map.put(toggle.getClass(), true);
    }
  }

  @Override
  public void prevent() {
    try {
      toggle.prevent();
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
      toggle.allow();
    } catch (UnsupportedOperationException ignored) {
      // Thrown when the security manager has been marked as unavailable
      // We have already warned the user about this in the call to `prevent`
      // so silently fall through
    }
  }

  private Unsafe getUnsafe() throws ReflectiveOperationException {
    // We want to avoid opening java.lang, so let's jump through hoops

    // First problem: we can't call `Unsafe.getUnsafe` directly and it might also be called
    // differently.
    Class<?> unsafe = Class.forName("sun.misc.Unsafe");
    for (Field f : unsafe.getDeclaredFields()) {
      if (f.getType() == unsafe) {
        f.setAccessible(true);
        return (Unsafe) f.get(null);
      }
    }
    throw new ReflectiveOperationException("Failed to get Unsafe");
  }
}

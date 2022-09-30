package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.lang.invoke.MethodHandles;
import java.util.Map;
import sun.misc.Unsafe;

public class Java17SystemExitToggle extends Java11SystemExitToggle {

  private SecurityManager previousSecurityManager;
  private TestRunningSecurityManager testingSecurityManager;

  public Java17SystemExitToggle() throws ReflectiveOperationException {
    suppressSecurityManagerWarning();
  }

  private void suppressSecurityManagerWarning() throws ReflectiveOperationException {
    Class<?> holderClazz = Class.forName("java.lang.System$CallersHolder");
    var callersField = holderClazz.getDeclaredField("callers");

    // We want to avoid opening java.lang, so let's jump through hoops

    // First problem: we can't call `Unsafe.getUnsafe` directly.
    var c = Class.forName("sun.misc.Unsafe");
    var lookup = MethodHandles.privateLookupIn(c, MethodHandles.lookup());
    var handle = lookup.findStaticVarHandle(c, "theUnsafe", c);
    var unsafe = (Unsafe) handle.get();

    // And we can't just grab the field easily
    var base = unsafe.staticFieldBase(callersField);
    long offset = unsafe.staticFieldOffset(callersField);
    Object callers = unsafe.getObject(base, offset);

    // And now we inject ourselves into the SecurityManager
    if (Map.class.isAssignableFrom(callers.getClass())) {
      @SuppressWarnings("unchecked")
      Map<Class<?>, Boolean> map = Map.class.cast(callers);
      map.put(getClass(), true);
    }
  }
}

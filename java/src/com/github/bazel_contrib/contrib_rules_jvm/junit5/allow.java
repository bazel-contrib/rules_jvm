import java.io.FileDescriptor;
import java.net.InetAddress;
import java.security.Permission;

/*
 * Prior to Java 18, it was always possible to replace the `SecurityManager`
 * at runtime. In Java 18, this changed, yet we still want to be able to
 * block calls to `System.exit` in tests that we execute.
 *
 * Fortunately, we can avoid the pain and grief by setting the
 * `java.security.manager` to `allow` in JDK 12+. Prior to JDK 12+, which
 * you will note includes the ever-so-popular Java 11, setting this setting
 * to `allow` will cause the SecurityManager to be set to a class who's
 * fully-qualified name is `allow`. Hence, this class.
 *
 * I can't apologise enough for this being here.
 */

// There is no package name: we want the fully qualifed name of this class to be `allow`
@SuppressWarnings({"deprecation", "removal"})
public class allow extends SecurityManager {

  @Override
  public void checkPermission(Permission perm) {}

  @Override
  public void checkPermission(Permission perm, Object context) {}

  @Override
  public void checkCreateClassLoader() {}

  @Override
  public void checkAccess(Thread t) {}

  @Override
  public void checkAccess(ThreadGroup g) {}

  @Override
  public void checkExit(int status) {}

  @Override
  public void checkExec(String cmd) {}

  @Override
  public void checkLink(String lib) {}

  @Override
  public void checkRead(FileDescriptor fd) {}

  @Override
  public void checkRead(String file) {}

  @Override
  public void checkRead(String file, Object context) {}

  @Override
  public void checkWrite(FileDescriptor fd) {}

  @Override
  public void checkWrite(String file) {}

  @Override
  public void checkDelete(String file) {}

  @Override
  public void checkConnect(String host, int port) {}

  @Override
  public void checkConnect(String host, int port, Object context) {}

  @Override
  public void checkListen(int port) {}

  @Override
  public void checkAccept(String host, int port) {}

  @Override
  public void checkMulticast(InetAddress maddr) {}

  @Override
  public void checkMulticast(InetAddress maddr, byte ttl) {}

  @Override
  public void checkPropertiesAccess() {}

  @Override
  public void checkPropertyAccess(String key) {}

  @Override
  public void checkPrintJobAccess() {}

  @Override
  public void checkPackageAccess(String pkg) {}

  @Override
  public void checkPackageDefinition(String pkg) {}

  @Override
  public void checkSetFactory() {}

  @Override
  public void checkSecurityAccess(String target) {}
}

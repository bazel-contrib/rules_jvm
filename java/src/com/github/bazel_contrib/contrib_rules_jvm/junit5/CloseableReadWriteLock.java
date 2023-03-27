package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.util.concurrent.locks.ReentrantReadWriteLock;

public class CloseableReadWriteLock {
  private final ReentrantReadWriteLock lock = new ReentrantReadWriteLock();

  public interface SilentAutoCloseable extends AutoCloseable {
    @Override
    void close();
  }

  public SilentAutoCloseable readLock() {
    lock.readLock().lock();
    return lock.readLock()::unlock;
  }

  public SilentAutoCloseable writeLock() {
    lock.writeLock().lock();
    return lock.writeLock()::unlock;
  }
}

package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.util.Collections;
import java.util.Optional;
import java.util.Set;
import org.junit.platform.engine.TestDescriptor;
import org.junit.platform.engine.TestSource;
import org.junit.platform.engine.TestTag;
import org.junit.platform.engine.UniqueId;

public class StubbedTestDescriptor implements TestDescriptor {

  private final UniqueId uniqueId;
  private final Type type;
  private final Optional<TestSource> source;
  private final Optional<TestDescriptor> parent;

  public StubbedTestDescriptor(UniqueId uniqueId) {
    this(uniqueId, Type.TEST);
  }

  public StubbedTestDescriptor(UniqueId uniqueId, Type type) {
    this(uniqueId, type, null);
  }

  public StubbedTestDescriptor(UniqueId uniqueId, Type type, TestSource source) {
    this(uniqueId, type, source, null);
  }

  public StubbedTestDescriptor(
      UniqueId uniqueId, Type type, TestSource source, TestDescriptor parent) {
    this.uniqueId = uniqueId;
    this.type = type;
    this.source = Optional.ofNullable(source);
    this.parent = Optional.ofNullable(parent);
  }

  @Override
  public UniqueId getUniqueId() {
    return uniqueId;
  }

  @Override
  public String getDisplayName() {
    return uniqueId.toString();
  }

  @Override
  public Set<TestTag> getTags() {
    return Collections.emptySet();
  }

  @Override
  public Optional<TestSource> getSource() {
    return source;
  }

  @Override
  public Optional<TestDescriptor> getParent() {
    return parent;
  }

  @Override
  public void setParent(TestDescriptor parent) {
    // Do nothing
  }

  @Override
  public Set<? extends TestDescriptor> getChildren() {
    return Collections.emptySet();
  }

  @Override
  public void addChild(TestDescriptor descriptor) {
    // Do nothing
  }

  @Override
  public void removeChild(TestDescriptor descriptor) {
    // Do nothing
  }

  @Override
  public void removeFromHierarchy() {
    // Do nothing
  }

  @Override
  public Type getType() {
    return type;
  }

  @Override
  public Optional<? extends TestDescriptor> findByUniqueId(UniqueId uniqueId) {
    return Optional.empty();
  }
}

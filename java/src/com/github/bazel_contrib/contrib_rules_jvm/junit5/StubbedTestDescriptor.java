package com.github.bazel_contrib.contrib_rules_jvm.junit5;

import java.util.Optional;
import java.util.Set;
import org.junit.platform.engine.TestDescriptor;
import org.junit.platform.engine.TestSource;
import org.junit.platform.engine.TestTag;
import org.junit.platform.engine.UniqueId;

public class StubbedTestDescriptor implements TestDescriptor {

  private final UniqueId uniqueId;

  public StubbedTestDescriptor(UniqueId uniqueId) {
    this.uniqueId = uniqueId;
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
    return Set.of();
  }

  @Override
  public Optional<TestSource> getSource() {
    return Optional.empty();
  }

  @Override
  public Optional<TestDescriptor> getParent() {
    return Optional.empty();
  }

  @Override
  public void setParent(TestDescriptor parent) {
    // Do nothin
  }

  @Override
  public Set<? extends TestDescriptor> getChildren() {
    return Set.of();
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
    return Type.TEST;
  }

  @Override
  public Optional<? extends TestDescriptor> findByUniqueId(UniqueId uniqueId) {
    return Optional.empty();
  }
}

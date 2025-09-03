package com.example;

import org.apache.logging.log4j.core.LogEvent;
import org.apache.logging.log4j.core.appender.AbstractAppender;
import org.apache.logging.log4j.core.config.plugins.Plugin;
import org.apache.logging.log4j.core.config.plugins.PluginFactory;

@Plugin(name = "LogProcessor")
public class LogProcessor extends AbstractAppender {
  @PluginFactory
  public static LogProcessor create() {
    return new LogProcessor();
  }

  @Override
  public void append(LogEvent event) {
    // This is not a very useful log appender, and that's ok.
  }
}

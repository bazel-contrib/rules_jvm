package workspace.com.gazelle.java.javaparser.generators;

import example.external.NeedsExporting;
import example.external.Outer;

public interface ExportingInterface {
  NeedsExporting get();

  Outer.Inner getInner();
}

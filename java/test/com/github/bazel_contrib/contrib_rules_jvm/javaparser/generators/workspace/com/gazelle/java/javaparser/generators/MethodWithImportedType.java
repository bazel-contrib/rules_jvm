package workspace.com.gazelle.java.javaparser.generators;

import com.example.OtherOuterReturnType;
import com.example.Outer.InnerReturnType;
import com.example.OuterReturnType;

public interface MethodWithImportedType {
  public OuterReturnType getOne();

  public OtherOuterReturnType.Inner getTwo();

  public InnerReturnType getThree();
}

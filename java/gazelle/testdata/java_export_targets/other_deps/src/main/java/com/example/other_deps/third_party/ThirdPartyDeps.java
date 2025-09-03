package com.example.other_deps.third_party;

import java.nio.file.Paths;
import com.google.common.collect.ImmutableList;

public class ThirdPartyDeps {
  public static String module() {
    return String.format(
        "ThirdParty(stdlib=%s, thirdParty=%s)",
        Paths.get("/tmp/path"),
        ImmutableList.of("Hello", "World"));
  }
}

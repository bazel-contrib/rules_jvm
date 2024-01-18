package com.example.hello;

import com.google.devtools.build.runfiles.AutoBazelRepository;
import com.google.devtools.build.runfiles.Runfiles;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;

@AutoBazelRepository
public class Hello {
    public static String sayHi() throws IOException {
        Runfiles.Preloaded runfiles = Runfiles.preload();
        String path = runfiles
                .withSourceRepository(AutoBazelRepository_Hello.NAME)
                .rlocation("runfiles_example/src/main/java/com/example/hello/data.txt");
        String fileContents = Files.readString(Path.of(path), StandardCharsets.UTF_8);

        return String.format("Hello %s", fileContents);
    }
}

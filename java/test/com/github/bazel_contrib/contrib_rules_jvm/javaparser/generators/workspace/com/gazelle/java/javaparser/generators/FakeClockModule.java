package workspace.com.gazelle.java.javaparser.generators;

public class FakeClockModule {
    private final Clock clock;
    private final FakeClock fakeClock;

    public FakeClockModule() {
        this.fakeClock = new FakeClock();
        this.clock = fakeClock;
    }
}

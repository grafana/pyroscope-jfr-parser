import jdk.jfr.*;
import java.nio.file.*;
import java.time.Duration;

public class ThreadSleepGenerator {
    public static void main(String[] args) throws Exception {
        Configuration config = Configuration.getConfiguration("default");
        Recording recording = new Recording(config);
        recording.enable("jdk.ThreadSleep").withThreshold(Duration.ZERO);
        recording.start();
        for (int i = 0; i < 10; i++) {
            Thread.sleep(50);
        }
        recording.stop();
        recording.dump(Path.of("threadsleep.jfr"));
        recording.close();
    }
}

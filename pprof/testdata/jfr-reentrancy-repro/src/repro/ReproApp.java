package repro;

import jdk.jfr.EventSettings;
import jdk.jfr.Recording;

import java.io.FileOutputStream;
import java.io.OutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.util.Arrays;

public final class ReproApp {
    public static void main(String[] args) throws Exception {
        if (args.length != 1) {
            throw new IllegalArgumentException("expected output recording path");
        }

        Path recordingPath = Path.of(args[0]).toAbsolutePath();
        Files.createDirectories(recordingPath.getParent());

        String target = "repro-" + "x".repeat(117) + ".dat";

        byte[] payload = new byte[4096];
        Arrays.fill(payload, (byte) 'x');

        if (BitsReentrancyAgent.wasTriggered()) {
            throw new IllegalStateException("Bits was transformed before the recording started");
        }

        try (Recording recording = new Recording()) {
            enable(recording, "jdk.FileWrite", true);
            enable(recording, "jdk.JavaExceptionThrow", true);
            enable(recording, "jdk.ObjectAllocationOutsideTLAB", true);

            recording.setToDisk(true);
            recording.start();

            if (BitsReentrancyAgent.wasTriggered()) {
                throw new IllegalStateException("Bits was transformed during recording startup");
            }

            for (int i = 0; i < 32; i++) {
                writeWithDepth(32, target, payload);
            }

            Thread.sleep(200);

            if (!BitsReentrancyAgent.wasTriggered()) {
                throw new IllegalStateException("Bits was never transformed");
            }

            recording.stop();
            recording.dump(recordingPath);
        }
    }

    private static void enable(Recording recording, String eventName, boolean stackTrace) {
        EventSettings settings = recording.enable(eventName).withThreshold(Duration.ZERO);
        if (stackTrace) {
            settings.withStackTrace();
        }
    }

    private static void writeWithDepth(int depth, String path, byte[] payload) throws Exception {
        if (depth == 0) {
            try (OutputStream outputStream = new FileOutputStream(path, true)) {
                outputStream.write(payload);
                outputStream.flush();
            }
            return;
        }
        writeWithDepth(depth - 1, path, payload);
    }

    private ReproApp() {
    }
}

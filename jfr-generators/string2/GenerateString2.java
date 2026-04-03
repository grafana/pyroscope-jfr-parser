import jdk.jfr.*;
import jdk.jfr.consumer.RecordingFile;
import jdk.jfr.consumer.RecordedEvent;

import java.nio.file.Path;
import java.time.Duration;
import java.util.Set;

/**
 * Generates a JFR file that contains strings with encoding 2 (CONSTANT_POOL reference).
 *
 * The JDK EventWriter.putString() uses encoding 2 when:
 * - String length > 128: pooled on first use
 * - String length 17-128: pooled after second occurrence
 *
 * This generator creates custom events with a SettingControl whose value is
 * a long string (>128 chars). These create jdk.ActiveSetting events with
 * inline string fields (name, value) where the value uses encoding 2.
 *
 * Usage:
 *   javac GenerateString2.java
 *   java GenerateString2
 *
 * Output: string2.jfr in the current directory.
 */
public class GenerateString2 {

    /**
     * Custom SettingControl that accepts arbitrary string values.
     */
    static class LongStringSetting extends SettingControl {
        private String value = "";

        @Override
        public String combine(Set<String> values) {
            return values.isEmpty() ? value : values.iterator().next();
        }

        @Override
        public void setValue(String value) {
            this.value = value;
        }

        @Override
        public String getValue() {
            return value;
        }
    }

    @Name("com.example.SampleEvent1")
    @Label("Sample Event 1")
    @StackTrace(true)
    static class SampleEvent1 extends Event {
        @Label("Message")
        String message;

        @SettingDefinition
        @Name("customFilter")
        public boolean customFilter(LongStringSetting setting) {
            return true;
        }
    }

    @Name("com.example.SampleEvent2")
    @Label("Sample Event 2")
    @StackTrace(true)
    static class SampleEvent2 extends Event {
        @Label("Message")
        String message;

        @SettingDefinition
        @Name("customFilter")
        public boolean customFilter(LongStringSetting setting) {
            return true;
        }
    }

    @Name("com.example.SampleEvent3")
    @Label("Sample Event 3")
    @StackTrace(true)
    static class SampleEvent3 extends Event {
        @Label("Message")
        String message;

        @SettingDefinition
        @Name("customFilter")
        public boolean customFilter(LongStringSetting setting) {
            return true;
        }
    }

    // Busywork to generate CPU profiling events
    static volatile long sink;

    static void doWork() {
        long s = 0;
        for (int i = 0; i < 10_000_000; i++) {
            s += i;
        }
        sink = s;
    }

    public static void main(String[] args) throws Exception {
        Path output = Path.of("string2.jfr");

        // Long value (>128 chars) will be pooled immediately on first use -> encoding 2
        String longValue = "com.example.filter.this.is.a.very.long.setting.value.that.exceeds."
                + "one.hundred.twenty.eight.characters.to.ensure.immediate.string.pool.encoding"
                + ".two.in.jfr.format.padding.padding.padding";

        // Use "profile" configuration for maximum event coverage
        Configuration config = Configuration.getConfiguration("profile");

        try (Recording recording = new Recording(config)) {
            // Enable CPU profiling
            recording.enable("jdk.ExecutionSample").withPeriod(Duration.ofMillis(10));
            recording.enable("jdk.ObjectAllocationSample").withPeriod(Duration.ofMillis(1));
            recording.enable("jdk.ActiveSetting");
            recording.enable("jdk.ActiveRecording");

            // Enable custom events with long setting values.
            // Each creates jdk.ActiveSetting events where the value field
            // is written via EventWriter.putString() with encoding 2.
            recording.enable(SampleEvent1.class)
                    .with("customFilter", longValue);
            recording.enable(SampleEvent2.class)
                    .with("customFilter", longValue);
            recording.enable(SampleEvent3.class)
                    .with("customFilter", longValue);

            recording.setMaxSize(10 * 1024 * 1024);
            recording.start();

            // Generate CPU work
            for (int round = 0; round < 20; round++) {
                doWork();
                for (int i = 0; i < 1000; i++) {
                    byte[] b = new byte[1024];
                    sink += b.length;
                }

                // Emit custom events
                for (int i = 0; i < 20; i++) {
                    SampleEvent1 e = new SampleEvent1();
                    e.message = "custom-message-that-exceeds-sixteen-chars-" + (i % 3);
                    e.commit();
                }

                Thread.sleep(50);
            }

            recording.stop();
            recording.dump(output);
        }

        System.out.println("Generated " + output.toAbsolutePath());

        // Check for long string values in ActiveSetting events
        try (RecordingFile rf = new RecordingFile(output)) {
            int total = 0;
            int activeSettings = 0;
            int longValues = 0;
            while (rf.hasMoreEvents()) {
                RecordedEvent event = rf.readEvent();
                total++;
                if (event.getEventType().getName().equals("jdk.ActiveSetting")) {
                    activeSettings++;
                    String val = event.getString("value");
                    if (val != null && val.length() > 16) {
                        longValues++;
                        if (longValues <= 5) {
                            System.out.println("ActiveSetting long value (len=" + val.length() +
                                    "): " + (val.length() > 60 ? val.substring(0, 60) + "..." : val));
                        }
                    }
                }
            }
            System.out.println("Total: " + total + " events, " +
                    activeSettings + " ActiveSetting, " + longValues + " with value>16 chars");
        }
    }
}

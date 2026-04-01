package repro;

import java.lang.instrument.ClassFileTransformer;
import java.lang.instrument.Instrumentation;
import java.security.ProtectionDomain;
import java.util.concurrent.atomic.AtomicBoolean;

public final class BitsReentrancyAgent {
    private static final AtomicBoolean triggered = new AtomicBoolean();

    public static void premain(String agentArgs, Instrumentation instrumentation) {
        instrumentation.addTransformer(new BitsTransformer(), false);
    }

    public static boolean wasTriggered() {
        return triggered.get();
    }

    private static final class BitsTransformer implements ClassFileTransformer {
        @Override
        public byte[] transform(
                Module module,
                ClassLoader loader,
                String className,
                Class<?> classBeingRedefined,
                ProtectionDomain protectionDomain,
                byte[] classfileBuffer
        ) {
            if (!"jdk/jfr/internal/Bits".equals(className) || !triggered.compareAndSet(false, true)) {
                return null;
            }

            try {
                // Mimic TTL's caught transform-time exception path. The exception is swallowed,
                // but JFR still records a JavaExceptionThrow event while the outer event is open.
                throw new IllegalStateException("intentional transform-time exception for " + className);
            } catch (IllegalStateException ignored) {
                // Intentionally ignored.
            }

            // Make a large allocation to encourage an ObjectAllocationOutsideTLAB event
            // without introducing any recursive class transformations.
            byte[] outsideTlab = new byte[2 * 1024 * 1024];
            outsideTlab[0] = 1;
            if (outsideTlab[0] == 2) {
                throw new AssertionError("unreachable");
            }
            return null;
        }
    }

    private BitsReentrancyAgent() {
    }
}

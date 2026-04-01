package pprof

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/grafana/jfr-parser/parser"
	"github.com/stretchr/testify/require"
)

func TestParseReentrancyRepro(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping JVM integration test in short mode")
	}

	for _, bin := range []string{"java", "javac", "jar"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not available: %v", bin, err)
		}
	}

	reproSources := filepath.Join("testdata", "jfr-reentrancy-repro", "src", "repro")
	javaSources, err := filepath.Glob(filepath.Join(reproSources, "*.java"))
	require.NoError(t, err)
	require.NotEmpty(t, javaSources, "missing Java reproduction sources in %s", reproSources)

	workDir := t.TempDir()
	classesDir := filepath.Join(workDir, "classes")
	require.NoError(t, os.Mkdir(classesDir, 0o755))

	javacArgs := append([]string{"-d", classesDir}, javaSources...)
	runCommand(t, "", "javac", javacArgs...)

	manifestPath := filepath.Join(workDir, "manifest.mf")
	manifest := "Manifest-Version: 1.0\nPremain-Class: repro.BitsReentrancyAgent\n\n"
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0o644))

	agentJar := filepath.Join(workDir, "repro-agent.jar")
	runCommand(t, "", "jar", "--create", "--file", agentJar, "--manifest", manifestPath, "-C", classesDir, ".")

	var parseErr error
	var recordingPath string
	var debugPath string
	for attempt := 1; attempt <= 3; attempt++ {
		recordingPath = filepath.Join(workDir, fmt.Sprintf("reentrancy-%d.jfr", attempt))
		debugPath = filepath.Join(workDir, fmt.Sprintf("reentrancy-%d.debug.txt", attempt))

		runCommand(
			t,
			workDir,
			"java",
			"-XX:FlightRecorderOptions=stackdepth=256",
			"-javaagent:"+agentJar,
			"-cp", agentJar,
			"repro.ReproApp",
			recordingPath,
		)

		parseErr = parseRecording(recordingPath, debugPath)
		if parseErr != nil {
			break
		}
	}

	require.Error(t, parseErr, "expected a corrupted JFR event, parser reached EOF")
	require.Contains(t, parseErr.Error(), "invalid event size")
	t.Logf("reproduced parser failure: %v", parseErr)
	t.Logf("recording: %s", recordingPath)
	t.Logf("debug trace: %s", debugPath)
}

func parseRecording(path, debugPath string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	debugFile, err := os.Create(debugPath)
	if err != nil {
		return err
	}
	defer debugFile.Close()

	p := parser.NewParser(data, parser.Options{
		SymbolProcessor: parser.ProcessSymbols,
	})
	p.DebugFile = debugFile

	total := 0
	for {
		_, err := p.ParseEvent()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(debugFile, "parsed %d events with EOF\n", total)
				return nil
			}
			fmt.Fprintf(debugFile, "ERROR after %d events: %v\n", total, err)
			return err
		}
		total++
	}
}

func runCommand(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("%s timed out: %s %v\n%s", name, name, args, out)
	}
	require.NoError(t, err, "%s failed: %s %v\n%s", name, name, args, out)
	return out
}

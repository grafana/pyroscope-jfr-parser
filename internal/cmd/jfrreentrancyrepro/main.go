package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/grafana/jfr-parser/parser"
)

//go:embed java/repro/*.java
var javaSources embed.FS

type command struct {
	out         string
	debug       string
	attempts    int
	keepWorkDir bool
}

func main() {
	c := parseCommand()
	if err := run(c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseCommand() command {
	var c command
	flag.StringVar(&c.out, "out", "", "destination for the corrupted JFR")
	flag.StringVar(&c.debug, "debug", "", "destination for parser debug output")
	flag.IntVar(&c.attempts, "attempts", 3, "number of JVM recording attempts before failing")
	flag.BoolVar(&c.keepWorkDir, "keep-workdir", false, "keep the temporary Java build directory")
	flag.Parse()
	return c
}

func run(c command) error {
	if c.attempts <= 0 {
		return fmt.Errorf("attempts must be positive")
	}
	if c.out == "" {
		c.out = filepath.Join("pprof", "testdata", "jfr-reentrancy-repro", "corrupted-recording.jfr")
	}
	if c.debug == "" {
		c.debug = defaultDebugPath(c.out)
	}
	if err := os.MkdirAll(filepath.Dir(c.out), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.debug), 0o755); err != nil {
		return err
	}
	for _, bin := range []string{"java", "javac", "jar"} {
		if _, err := exec.LookPath(bin); err != nil {
			return fmt.Errorf("%s not available: %w", bin, err)
		}
	}

	workDir, err := os.MkdirTemp("", "jfr-reentrancy-repro-")
	if err != nil {
		return err
	}
	if c.keepWorkDir {
		fmt.Printf("keeping work dir: %s\n", workDir)
	} else {
		defer os.RemoveAll(workDir)
	}

	srcDir := filepath.Join(workDir, "src")
	classesDir := filepath.Join(workDir, "classes")
	if err := os.MkdirAll(classesDir, 0o755); err != nil {
		return err
	}
	javaFiles, err := writeEmbeddedSources(srcDir)
	if err != nil {
		return err
	}

	javacArgs := append([]string{"-d", classesDir}, javaFiles...)
	if _, err := runCommand("", "javac", javacArgs...); err != nil {
		return err
	}

	manifestPath := filepath.Join(workDir, "manifest.mf")
	manifest := "Manifest-Version: 1.0\nPremain-Class: repro.BitsReentrancyAgent\n\n"
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		return err
	}

	agentJar := filepath.Join(workDir, "repro-agent.jar")
	if _, err := runCommand("", "jar", "--create", "--file", agentJar, "--manifest", manifestPath, "-C", classesDir, "."); err != nil {
		return err
	}

	for attempt := 1; attempt <= c.attempts; attempt++ {
		recordingPath := filepath.Join(workDir, fmt.Sprintf("reentrancy-%d.jfr", attempt))
		debugPath := filepath.Join(workDir, fmt.Sprintf("reentrancy-%d.debug.txt", attempt))

		if _, err := runCommand(
			workDir,
			"java",
			"-XX:FlightRecorderOptions=stackdepth=256",
			"-javaagent:"+agentJar,
			"-cp", agentJar,
			"repro.ReproApp",
			recordingPath,
		); err != nil {
			return fmt.Errorf("recording attempt %d failed: %w", attempt, err)
		}

		parseErr, total, err := parseRecording(recordingPath, debugPath)
		if err != nil {
			return err
		}
		if parseErr == nil {
			fmt.Printf("attempt %d reached EOF after %d events\n", attempt, total)
			continue
		}

		if err := copyFile(recordingPath, c.out); err != nil {
			return err
		}
		if err := copyFile(debugPath, c.debug); err != nil {
			return err
		}

		fmt.Printf("corruption reproduced on attempt %d\n", attempt)
		fmt.Printf("recording: %s\n", c.out)
		fmt.Printf("debug: %s\n", c.debug)
		fmt.Printf("parser error after %d events: %v\n", total, parseErr)
		return nil
	}

	return errors.New("parser reached EOF in all attempts; corruption not reproduced")
}

func writeEmbeddedSources(dstRoot string) ([]string, error) {
	matches, err := fs.Glob(javaSources, "java/repro/*.java")
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("no embedded Java sources found")
	}
	out := make([]string, 0, len(matches))
	for _, match := range matches {
		data, err := javaSources.ReadFile(match)
		if err != nil {
			return nil, err
		}
		target := filepath.Join(dstRoot, strings.TrimPrefix(match, "java/"))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return nil, err
		}
		out = append(out, target)
	}
	return out, nil
}

func parseRecording(path, debugPath string) (parseErr error, total int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}

	debugFile, err := os.Create(debugPath)
	if err != nil {
		return nil, 0, err
	}
	defer debugFile.Close()

	p := parser.NewParser(data, parser.Options{
		SymbolProcessor: parser.ProcessSymbols,
	})
	p.DebugFile = debugFile

	for {
		_, err := p.ParseEvent()
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(debugFile, "parsed %d events with EOF\n", total)
				return nil, total, nil
			}
			fmt.Fprintf(debugFile, "ERROR after %d events: %v\n", total, err)
			return err, total, nil
		}
		total++
	}
}

func runCommand(dir, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("%s timed out: %s %v\n%s", name, name, args, out)
	}
	if err != nil {
		return nil, fmt.Errorf("%s failed: %s %v\n%s", name, name, args, out)
	}
	return out, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func defaultDebugPath(recordingPath string) string {
	ext := filepath.Ext(recordingPath)
	if ext == "" {
		return recordingPath + ".debug.txt"
	}
	return strings.TrimSuffix(recordingPath, ext) + ".debug.txt"
}

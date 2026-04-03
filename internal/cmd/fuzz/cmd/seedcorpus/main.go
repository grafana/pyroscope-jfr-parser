package main

import (
	"compress/gzip"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func repoRoot() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git rev-parse --show-toplevel: %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(out))
}

func main() {
	rootDir := repoRoot()

	testdata := flag.String("testdata", filepath.Join(rootDir, "parser", "testdata"), "directory containing .jfr.gz test files")
	corpus := flag.String("corpus", "corpus", "output directory for decompressed corpus files")
	maxSize := flag.Int64("max-size", 524288, "maximum compressed file size in bytes")
	flag.Parse()

	if err := generate(*testdata, *corpus, *maxSize); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func generate(srcDir string, destDir string, maxCompressedSize int64) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", destDir, err)
	}

	matches, err := filepath.Glob(filepath.Join(srcDir, "*.jfr.gz"))
	if err != nil {
		return fmt.Errorf("glob %s: %w", srcDir, err)
	}

	for _, m := range matches {
		info, err := os.Stat(m)
		if err != nil {
			return fmt.Errorf("stat %s: %w", m, err)
		}
		if info.Size() > maxCompressedSize {
			continue
		}

		jfrData, err := readGzip(m)
		if err != nil {
			return err
		}

		labels, err := findLabels(m)
		if err != nil {
			return err
		}

		base := filepath.Base(m)
		name := strings.TrimSuffix(base, ".gz")
		outPath := filepath.Join(destDir, name)

		if err := os.WriteFile(outPath, encodeFuzzInput(jfrData, labels), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}
	return nil
}

func readGzip(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gzip reader %s: %w", path, err)
	}
	defer gr.Close()

	data, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("decompress %s: %w", path, err)
	}
	return data, nil
}

func findLabels(jfrGzPath string) ([]byte, error) {
	base := strings.TrimSuffix(jfrGzPath, ".jfr.gz")
	for _, suffix := range []string{".labels.pb.gz", ".labels.gz"} {
		path := base + suffix
		if _, err := os.Stat(path); err == nil {
			return readGzip(path)
		}
	}
	return nil, nil
}

func encodeFuzzInput(jfrData []byte, labels []byte) []byte {
	var flags byte
	if len(labels) > 0 {
		flags |= 1
	}
	buf := []byte{flags}
	if len(labels) > 0 {
		buf = append(buf, byte(len(labels)))
		buf = append(buf, labels...)
	}
	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, 1706241880000)
	buf = append(buf, ts...)
	binary.LittleEndian.PutUint64(ts, 1706241890000)
	buf = append(buf, ts...)
	binary.LittleEndian.PutUint64(ts, 100)
	buf = append(buf, ts...)
	return append(buf, jfrData...)
}

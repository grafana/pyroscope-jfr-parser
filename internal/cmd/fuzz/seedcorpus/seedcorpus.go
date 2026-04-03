package seedcorpus

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// extract decompresses a .jfr.gz file and returns the raw JFR bytes.
func extract(gzPath string) ([]byte, error) {
	f, err := os.Open(gzPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", gzPath, err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("gzip reader %s: %w", gzPath, err)
	}
	defer gr.Close()

	data, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("decompress %s: %w", gzPath, err)
	}
	return data, nil
}

// EncodeFuzzInput wraps raw JFR bytes into the binary format expected by
// LLVMFuzzerTestOneInput: flags(1B) + startTime(8B LE) + endTime(8B LE) +
// sampleRate(8B LE) + jfrData.
func EncodeFuzzInput(jfrData []byte) []byte {
	header := make([]byte, 1+8+8+8)
	header[0] = 0 // flags: no labels, no truncated frame
	binary.LittleEndian.PutUint64(header[1:9], 1706241880000)
	binary.LittleEndian.PutUint64(header[9:17], 1706241890000)
	binary.LittleEndian.PutUint64(header[17:25], 100)
	return append(header, jfrData...)
}

// Generate finds .jfr.gz files in srcDir that are at most maxCompressedSize
// bytes, decompresses them, wraps each in the fuzzer's binary input format,
// and writes them to destDir.
func Generate(srcDir string, destDir string, maxCompressedSize int64) error {
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

		jfrData, err := extract(m)
		if err != nil {
			return err
		}

		base := filepath.Base(m)
		name := strings.TrimSuffix(base, ".gz")
		outPath := filepath.Join(destDir, name)

		if err := os.WriteFile(outPath, EncodeFuzzInput(jfrData), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}
	return nil
}

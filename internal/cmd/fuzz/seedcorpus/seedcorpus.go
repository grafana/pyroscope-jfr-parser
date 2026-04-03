package seedcorpus

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extract decompresses a single .jfr.gz file into destDir.
// Returns the output file path.
func Extract(gzPath string, destDir string) (string, error) {
	f, err := os.Open(gzPath)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", gzPath, err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip reader %s: %w", gzPath, err)
	}
	defer gr.Close()

	base := filepath.Base(gzPath)
	name := strings.TrimSuffix(base, ".gz")
	outPath := filepath.Join(destDir, name)

	out, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", outPath, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, gr); err != nil {
		return "", fmt.Errorf("decompress %s: %w", gzPath, err)
	}
	return outPath, nil
}

// Generate finds .jfr.gz files in srcDir that are at most maxCompressedSize
// bytes and decompresses them into destDir using Extract.
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
		if _, err := Extract(m, destDir); err != nil {
			return err
		}
	}
	return nil
}

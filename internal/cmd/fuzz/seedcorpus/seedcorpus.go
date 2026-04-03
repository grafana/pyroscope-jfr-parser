package seedcorpus

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FuzzInput is the decoded fuzz input that the fuzzer operates on.
type FuzzInput struct {
	TruncatedFrame bool
	Labels         []byte
	StartTime      time.Time
	EndTime        time.Time
	SampleRate     int64
	JFR            []byte
}

// DecodeFuzzInput parses the binary format used by the fuzzer corpus.
// Format: flags(1B) + [labels_len(1B) + labels(NB)] + startTime(8B LE) +
// endTime(8B LE) + sampleRate(8B LE) + jfrData.
func DecodeFuzzInput(data []byte) FuzzInput {
	r := reader{data: data}
	flags := r.u8()
	withLabels := flags&1 == 1
	truncatedFrame := (flags>>1)&1 == 1

	var labels []byte
	if withLabels {
		labels = r.bytes(int(r.u8()))
	}

	return FuzzInput{
		TruncatedFrame: truncatedFrame,
		Labels:         labels,
		StartTime:      time.UnixMilli(int64(r.u64())),
		EndTime:        time.UnixMilli(int64(r.u64())),
		SampleRate:     int64(r.u64()),
		JFR:            r.rest(),
	}
}

type reader struct {
	data []byte
}

func (r *reader) u8() uint8 {
	if len(r.data) == 0 {
		return 0
	}
	b := r.data[0]
	r.data = r.data[1:]
	return b
}

func (r *reader) u64() uint64 {
	if len(r.data) < 8 {
		return 0
	}
	v := binary.LittleEndian.Uint64(r.data[:8])
	r.data = r.data[8:]
	return v
}

func (r *reader) bytes(sz int) []byte {
	if sz == 0 {
		return nil
	}
	if len(r.data) < sz {
		res := r.data
		r.data = nil
		return res
	}
	res := r.data[:sz]
	r.data = r.data[sz:]
	return res
}

func (r *reader) rest() []byte {
	res := r.data
	r.data = nil
	return res
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

// findLabels looks for a labels file adjacent to a .jfr.gz file.
// Checks for <name>.labels.pb.gz and <name>.labels.gz.
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

// Generate finds .jfr.gz files in srcDir that are at most maxCompressedSize
// bytes, decompresses them, wraps each in the fuzzer's binary input format
// (including adjacent labels if present), and writes them to destDir.
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

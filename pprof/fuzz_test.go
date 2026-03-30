package pprof

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func FuzzParseJFR(f *testing.F) {
	// Seed from fuzz-corpus/ directory (may not exist — that's fine).
	entries, _ := os.ReadDir(filepath.Join("..", "fuzz-corpus"))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join("..", "fuzz-corpus", e.Name()))
		if err != nil {
			continue
		}
		f.Add(data)
	}

	pi := &ParseInput{
		StartTime:  time.Unix(1706241880, 0),
		EndTime:    time.Unix(1706241890, 0),
		SampleRate: 100,
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseJFR(data, pi, nil, WithDisablePanicRecovery(true))
	})
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/grafana/jfr-parser/internal/cmd/fuzz/seedcorpus"
)

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile is internal/cmd/fuzz/seedcorpus/cmd/main.go
	// repo root is 5 levels up
	rootDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "..")

	testdata := flag.String("testdata", filepath.Join(rootDir, "parser", "testdata"), "directory containing .jfr.gz test files")
	corpus := flag.String("corpus", "corpus", "output directory for decompressed corpus files")
	maxSize := flag.Int64("max-size", 524288, "maximum compressed file size in bytes")
	flag.Parse()

	if err := seedcorpus.Generate(*testdata, *corpus, *maxSize); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

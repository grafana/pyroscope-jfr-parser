package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/grafana/jfr-parser/internal/corpus"
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
	corpusDir := flag.String("corpus", filepath.Join(rootDir, "internal", "cmd", "fuzz", "corpus"), "output directory for corpus files")
	maxSize := flag.Int64("max-size", 524288, "maximum compressed file size in bytes")
	flag.Parse()

	if err := corpus.Generate(*testdata, *corpusDir, *maxSize); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/grafana/jfr-parser/internal/cmd/fuzz/seedcorpus"
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

	if err := seedcorpus.Generate(*testdata, *corpus, *maxSize); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

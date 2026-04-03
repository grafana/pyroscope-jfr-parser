package main

import (
	"fmt"
	"os"

	"github.com/grafana/jfr-parser/internal/corpus"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: fuzzreplay <file> [file...]\n")
		os.Exit(1)
	}

	for _, path := range os.Args[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		_, err = corpus.ParseOne(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		} else {
			fmt.Fprintf(os.Stderr, "%s: ok\n", path)
		}
	}
}

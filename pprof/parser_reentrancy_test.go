package pprof

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/grafana/jfr-parser/parser"
)

func TestParseReentrancyRepro(t *testing.T) {
	const path = "/home/korniltsev/jfr-reentrancy-repro/recording.jfr"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skip("test file not available:", err)
	}

	debugFile, err := os.Create("/tmp/jfr-reentrancy.txt")
	if err != nil {
		t.Fatal(err)
	}

	p := parser.NewParser(data, parser.Options{
		SymbolProcessor: parser.ProcessSymbols,
	})
	p.DebugFile = debugFile
	defer func() {
		fmt.Fprintf(debugFile, "file=%s\n", path)
		debugFile.Close()
	}()

	total := 0
	errors := 0
	for {
		_, err := p.ParseEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			errors++
			t.Logf("ParseEvent error after %d events: %v", total, err)
			break
		}
		total++
	}

	t.Logf("parsed %d events, %d errors", total, errors)
	if errors > 0 {
		t.Logf("CORRUPTION REPRODUCED!")
	}
}

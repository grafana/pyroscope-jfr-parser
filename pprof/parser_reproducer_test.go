package pprof

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/grafana/jfr-parser/parser"
)

func TestParseReproducer(t *testing.T) {
	const path = "/tmp/reproducer.jfr"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skip("test file not available:", err)
	}

	debugFile, err := os.Create("/tmp/jfr-reproducer-events.txt")
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
			fmt.Fprintf(debugFile, "ERROR after %d events: %v\n", total, err)
			t.Logf("ParseEvent error after %d events: %v", total, err)
			errors++
			break
		}
		total++
	}

	t.Logf("parsed %d events, %d errors", total, errors)
	if errors > 0 {
		t.Logf("CORRUPTION REPRODUCED!")
	} else {
		t.Logf("no corruption detected")
	}
}

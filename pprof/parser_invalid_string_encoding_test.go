package pprof

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/grafana/jfr-parser/parser"
	"github.com/grafana/jfr-parser/parser/types/def"
)

func TestParseInvalidStringEncoding(t *testing.T) {
	//const path = "/home/korniltsev/failed-jfrs-fail/836243/1774887701945466763-7bb1ce24-13d9-49f8-a4e9-1ebdcfceee1c.jfr"
	const path = "/home/korniltsev/failed-jfrs/836243/1774887701945466763-7bb1ce24-13d9-49f8-a4e9-1ebdcfceee1c.jfr"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skip("test file not available:", err)
	}

	debugFile, err := os.Create("/tmp/jfr-events.txt")
	if err != nil {
		t.Fatal(err)
	}
	p := parser.NewParser(data, parser.Options{
		SymbolProcessor: parser.ProcessSymbols,
	})
	p.DebugFile = debugFile
	p.DebugTraceAll = true
	defer func() {
		fmt.Fprintf(debugFile, "file=%s\n", path)
		debugTypes := []string{
			"jdk.JavaExceptionThrow",
			"jdk.FileWrite",
			"jdk.InitialSystemProperty",
			"jdk.MetaspaceChunkFreeListSummary",
			"jdk.OldGarbageCollection",
			"jdk.NetworkUtilization",
			"jdk.DataLoss",
		}
		for _, name := range debugTypes {
			if cls, ok := p.TypeMap.NameMap[name]; ok {
				fmt.Fprintf(debugFile, "%s\n", cls)
			}
		}
		// Print referenced types
		refTypes := []int{231, 232, 239, 188, 212, 190, 197, 202, 187}
		for _, id := range refTypes {
			if cls, ok := p.TypeMap.IDMap[def.TypeID(id)]; ok {
				fmt.Fprintf(debugFile, "type %d = %s\n", id, cls)
			} else {
				fmt.Fprintf(debugFile, "type %d = NOT FOUND\n", id)
			}
		}
		debugFile.Close()
	}()

	counts := make(map[string]int)
	total := 0
	for {
		typ, err := p.ParseEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("ParseEvent failed after %d events: %v", total, err)
		}
		total++
		if cls, ok := p.TypeMap.IDMap[typ]; ok {
			counts[cls.Name]++
		}
	}

	t.Logf("parsed %d event types", len(counts))
	for name, count := range counts {
		t.Logf("  %s: %d", name, count)
	}
}

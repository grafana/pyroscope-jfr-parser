package pprof

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/jfr-parser/parser"
	"github.com/grafana/jfr-parser/parser/types/def"
)

func TestScanAllJFRFiles(t *testing.T) {
	const dir = "/home/korniltsev/failed-jfrs/836243"
	files, err := filepath.Glob(filepath.Join(dir, "*.jfr"))
	if err != nil || len(files) == 0 {
		t.Skip("no JFR files found in", dir)
	}

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Logf("skip %s: %v", path, err)
			continue
		}

		debugPath := strings.TrimSuffix(path, ".jfr") + ".debug.txt"
		debugFile, err := os.Create(debugPath)
		if err != nil {
			t.Fatalf("create debug file %s: %v", debugPath, err)
		}

		p := parser.NewParser(data, parser.Options{
			SymbolProcessor: parser.ProcessSymbols,
		})
		p.DebugFile = debugFile

		fmt.Fprintf(debugFile, "=== FILE: %s (size=%d) ===\n", filepath.Base(path), len(data))

		total := 0
		jvmInfoPrinted := false
		for {
			typ, err := p.ParseEvent()
			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Fprintf(debugFile, "ERROR after %d events: %v\n", total, err)
				break
			}
			total++
			if !jvmInfoPrinted {
				if cls, ok := p.TypeMap.IDMap[typ]; ok && cls.Name == "jdk.JVMInformation" {
					// won't happen - it's skipped. handled below
					_ = cls
				}
			}
		}

		// Print JVM info by scanning the type map for jdk.JVMInformation
		if jvmCls, ok := p.TypeMap.NameMap["jdk.JVMInformation"]; ok {
			printJVMInfo(debugFile, p, data, jvmCls)
		} else {
			fmt.Fprintf(debugFile, "jdk.JVMInformation type not found in metadata\n")
		}

		fmt.Fprintf(debugFile, "total events: %d\n", total)
		debugFile.Close()
	}
}

func printJVMInfo(w io.Writer, p *parser.Parser, data []byte, jvmCls *def.Class) {
	// Manual scan: iterate through events in the buffer looking for the JVMInformation type ID
	pos := 68 // skip chunk header
	bufLen := len(data)
	targetType := jvmCls.ID

	for pos < bufLen {
		startPos := pos
		// read size varlong
		size, n := readVarLong(data, pos)
		if n <= 0 || size == 0 {
			break
		}
		pos += n
		// read type varlong
		typ, n := readVarLong(data, pos)
		if n <= 0 {
			break
		}
		pos += n

		eventEnd := startPos + int(size)
		if eventEnd > bufLen || eventEnd <= startPos {
			break
		}

		if def.TypeID(typ) == targetType {
			// Read fields reflectively
			fmt.Fprintf(w, "jdk.JVMInformation at pos=%d:\n", startPos)
			for _, f := range jvmCls.Fields {
				if pos >= eventEnd {
					break
				}
				fieldType, _ := p.TypeMap.IDMap[f.Type]
				typeName := ""
				if fieldType != nil {
					typeName = fieldType.Name
				}
				if f.ConstantPool {
					v, n := readVarLong(data, pos)
					if n <= 0 {
						break
					}
					pos += n
					fmt.Fprintf(w, "  %s: ref=%d\n", f.Name, v)
				} else if typeName == "long" || typeName == "int" || typeName == "short" || typeName == "byte" || typeName == "boolean" {
					v, n := readVarLong(data, pos)
					if n <= 0 {
						break
					}
					pos += n
					fmt.Fprintf(w, "  %s: %d\n", f.Name, v)
				} else if typeName == "java.lang.String" {
					s, newPos := readString(data, pos)
					if newPos <= pos {
						break
					}
					pos = newPos
					fmt.Fprintf(w, "  %s: %q\n", f.Name, s)
				} else {
					fmt.Fprintf(w, "  %s: <skip, type=%s>\n", f.Name, typeName)
					break
				}
			}
			return
		}
		pos = eventEnd
	}
	fmt.Fprintf(w, "jdk.JVMInformation event not found\n")
}

func readVarLong(data []byte, pos int) (uint64, int) {
	v := uint64(0)
	n := 0
	for shift := uint(0); shift <= 56; shift += 7 {
		if pos+n >= len(data) {
			return 0, -1
		}
		b := data[pos+n]
		n++
		if shift == 56 {
			v |= uint64(b&0xFF) << shift
			break
		}
		v |= uint64(b&0x7F) << shift
		if b < 0x80 {
			break
		}
	}
	return v, n
}

func readString(data []byte, pos int) (string, int) {
	if pos >= len(data) {
		return "", pos
	}
	enc := data[pos]
	pos++
	switch enc {
	case 0, 1:
		return "", pos
	case 2:
		// constant pool ref - just read the key, can't resolve here
		_, n := readVarLong(data, pos)
		if n <= 0 {
			return "", pos
		}
		return "<cpref>", pos + n
	case 3:
		// latin1 byte array
		length, n := readVarLong(data, pos)
		if n <= 0 {
			return "", pos
		}
		pos += n
		end := pos + int(length)
		if end > len(data) || end < pos {
			return "", pos
		}
		return string(data[pos:end]), end
	case 4:
		// char array (varlong per char)
		length, n := readVarLong(data, pos)
		if n <= 0 {
			return "", pos
		}
		pos += n
		runes := make([]rune, int(length))
		for i := 0; i < int(length); i++ {
			v, n := readVarLong(data, pos)
			if n <= 0 {
				return string(runes[:i]), pos
			}
			runes[i] = rune(v)
			pos += n
		}
		return string(runes), pos
	default:
		return "", pos
	}
}

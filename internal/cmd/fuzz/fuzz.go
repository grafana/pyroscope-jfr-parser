//go:build libfuzzer

package main

// #include <stdint.h>
import "C"
import (
	"unsafe"

	"github.com/grafana/jfr-parser/internal/corpus"
	"github.com/grafana/jfr-parser/pprof"
)

//export LLVMFuzzerInitialize
func LLVMFuzzerInitialize(argc *C.int, argv ***C.char) C.int {
	return 0
}

//export LLVMFuzzerTestOneInput
func LLVMFuzzerTestOneInput(data *C.char, size C.size_t) C.int {
	gdata := unsafe.Slice((*byte)(unsafe.Pointer(data)), size)
	if len(gdata) == 0 {
		return 0
	}

	fi := corpus.Decode(gdata)
	_, _ = pprof.ParseJFR(fi.JFR, fi.ParseInput, fi.Labels, pprof.WithTruncatedFrame(fi.TruncatedFrame), pprof.WithDisablePanicRecovery(true))
	return 0
}

func main() {

}

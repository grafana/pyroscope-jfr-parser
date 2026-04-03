//go:build libfuzzer

package main

// #include <stdint.h>
import "C"
import (
	"github.com/grafana/jfr-parser/pprof"
	"unsafe"
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

	fi := decodeFuzzInput(gdata)
	_, _ = pprof.ParseJFR(fi.jfr, fi.parseInput, fi.labels, pprof.WithTruncatedFrame(fi.truncatedFrame), pprof.WithDisablePanicRecovery(true))
	return 0
}

func main() {

}

//go:build libfuzzer

package main

// #include <stdint.h>
import "C"
import (
	"unsafe"

	"github.com/grafana/jfr-parser/internal/corpus"
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
	_, _ = corpus.ParseOne(gdata)
	return 0
}

func main() {

}

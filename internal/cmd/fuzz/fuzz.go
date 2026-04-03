//go:build libfuzzer

package main

// #include <stdint.h>
import "C"
import (
	"github.com/grafana/jfr-parser/internal/cmd/fuzz/seedcorpus"
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

	fi := seedcorpus.DecodeFuzzInput(gdata)

	var ls *pprof.LabelsSnapshot
	if len(fi.Labels) > 0 {
		ls = &pprof.LabelsSnapshot{}
		_ = ls.UnmarshalVT(fi.Labels)
	}

	pi := &pprof.ParseInput{
		StartTime:  fi.StartTime,
		EndTime:    fi.EndTime,
		SampleRate: fi.SampleRate,
	}

	_, _ = pprof.ParseJFR(fi.JFR, pi, ls, pprof.WithTruncatedFrame(fi.TruncatedFrame), pprof.WithDisablePanicRecovery(true))
	return 0
}

func main() {

}

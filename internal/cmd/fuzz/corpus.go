package main

import (
	"encoding/binary"
	"time"

	"github.com/grafana/jfr-parser/pprof"
)

type fuzzInput struct {
	truncatedFrame bool
	labels         *pprof.LabelsSnapshot
	parseInput     *pprof.ParseInput
	jfr            []byte
}

func decodeFuzzInput(data []byte) fuzzInput {
	r := reader{data: data}
	flags := r.u8()
	withLabels := flags&1 == 1
	truncatedFrame := (flags>>1)&1 == 1

	var ls *pprof.LabelsSnapshot
	if withLabels {
		lsb := r.bytes(int(r.u8()))
		ls = &pprof.LabelsSnapshot{}
		_ = ls.UnmarshalVT(lsb)
	}

	return fuzzInput{
		truncatedFrame: truncatedFrame,
		labels:         ls,
		parseInput: &pprof.ParseInput{
			StartTime:  time.UnixMilli(int64(r.u64())),
			EndTime:    time.UnixMilli(int64(r.u64())),
			SampleRate: int64(r.u64()),
		},
		jfr: r.rest(),
	}
}

type reader struct {
	data []byte
}

func (r *reader) u8() uint8 {
	if len(r.data) == 0 {
		return 0
	}
	b := r.data[0]
	r.data = r.data[1:]
	return b
}

func (r *reader) u64() uint64 {
	if len(r.data) < 8 {
		return 0
	}
	v := binary.LittleEndian.Uint64(r.data[:8])
	r.data = r.data[8:]
	return v
}

func (r *reader) bytes(sz int) []byte {
	if sz == 0 {
		return nil
	}
	if len(r.data) < sz {
		res := r.data
		r.data = nil
		return res
	}
	res := r.data[:sz]
	r.data = r.data[sz:]
	return res
}

func (r *reader) rest() []byte {
	res := r.data
	r.data = nil
	return res
}

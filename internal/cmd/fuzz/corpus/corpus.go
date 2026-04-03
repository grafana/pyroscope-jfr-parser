package corpus

import (
	"encoding/binary"
	"time"

	"github.com/grafana/jfr-parser/pprof"
)

// Input is the decoded fuzz input that the fuzzer operates on.
type Input struct {
	TruncatedFrame bool
	Labels         *pprof.LabelsSnapshot
	ParseInput     *pprof.ParseInput
	JFR            []byte
}

// Decode parses the binary format used by the fuzzer corpus.
// Format: flags(1B) + [labels_len(1B) + labels(NB)] + startTime(8B LE) +
// endTime(8B LE) + sampleRate(8B LE) + jfrData.
func Decode(data []byte) Input {
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

	return Input{
		TruncatedFrame: truncatedFrame,
		Labels:         ls,
		ParseInput: &pprof.ParseInput{
			StartTime:  time.UnixMilli(int64(r.u64())),
			EndTime:    time.UnixMilli(int64(r.u64())),
			SampleRate: int64(r.u64()),
		},
		JFR: r.rest(),
	}
}

// Encode wraps raw JFR bytes and optional labels into the binary format
// expected by the fuzzer.
func Encode(jfrData []byte, labels []byte) []byte {
	var flags byte
	if len(labels) > 0 {
		flags |= 1
	}
	buf := []byte{flags}
	if len(labels) > 0 {
		buf = append(buf, byte(len(labels)))
		buf = append(buf, labels...)
	}
	ts := make([]byte, 8)
	binary.LittleEndian.PutUint64(ts, 1706241880000)
	buf = append(buf, ts...)
	binary.LittleEndian.PutUint64(ts, 1706241890000)
	buf = append(buf, ts...)
	binary.LittleEndian.PutUint64(ts, 100)
	buf = append(buf, ts...)
	return append(buf, jfrData...)
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

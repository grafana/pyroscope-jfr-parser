package pprof

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	gpprof "github.com/google/pprof/profile"
	profilev1 "github.com/grafana/pyroscope/api/gen/proto/go/google/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testdataDir = "../parser/testdata/"
const doDump = false

type testdata struct {
	jfr, labels   string
	testName      string
	expectedCount int
	options       []Option
}

var testFiles = []testdata{
	{
		jfr:           "example",
		labels:        "",
		expectedCount: 4,
		options:       nil,
	},

	{
		jfr:           "async-profiler",
		labels:        "",
		expectedCount: 3,
		options:       nil,
	},
	{
		jfr:           "goland",
		labels:        "",
		expectedCount: 5,
		options:       nil,
	},
	{
		jfr:           "goland-multichunk",
		labels:        "",
		expectedCount: 5,
		options:       nil,
	},
	{
		jfr:           "FastSlow_2024_01_16_180855",
		labels:        "",
		expectedCount: 3,
		options:       nil,
	},
	{
		jfr:           "cortex-dev-01__kafka-0__cpu__0",
		labels:        "",
		expectedCount: 1,
		options:       nil,
	},
	{
		jfr:           "cortex-dev-01__kafka-0__cpu__1",
		labels:        "",
		expectedCount: 1,
		options:       nil,
	},
	{
		jfr:           "cortex-dev-01__kafka-0__cpu__2",
		labels:        "",
		expectedCount: 1,
		options:       nil,
	},
	{
		jfr:           "cortex-dev-01__kafka-0__cpu__3",
		labels:        "",
		expectedCount: 1,
		options:       nil,
	},
	{
		jfr:           "cortex-dev-01__kafka-0__cpu_lock0_alloc0__0",
		labels:        "",
		expectedCount: 5,
		options:       nil,
	},
	{
		jfr:           "cortex-dev-01__kafka-0__cpu_lock_alloc__0",
		labels:        "",
		expectedCount: 2,
		options:       nil,
	},
	{
		jfr:           "cortex-dev-01__kafka-0__cpu_lock_alloc__1",
		labels:        "",
		expectedCount: 2,
		options:       nil,
	},
	{
		jfr:           "cortex-dev-01__kafka-0__cpu_lock_alloc__2",
		labels:        "",
		expectedCount: 2,
		options:       nil,
	},
	{
		jfr:           "cortex-dev-01__kafka-0__cpu_lock_alloc__3",
		labels:        "",
		expectedCount: 2,
		options:       nil,
	},
	{
		jfr:           "dump1",
		labels:        "dump1.labels.pb.gz",
		expectedCount: 1,
		options:       nil,
	},
	{
		jfr:           "dump2",
		labels:        "dump2.labels.pb.gz",
		expectedCount: 4,
		options:       nil,
	},
	{
		jfr:           "dd-trace-java",
		labels:        "",
		expectedCount: 4,
		options:       nil,
	},
	{
		jfr:           "cpool-uint64-constant-index",
		labels:        "",
		expectedCount: 5,
		options:       nil,
	},
	{
		jfr:           "event-with-type-zero",
		labels:        "",
		expectedCount: 5,
		options:       nil,
	},
	{
		jfr:           "event-with-type-zero",
		testName:      "event-with-type-zero with truncated frame",
		labels:        "",
		expectedCount: 5,
		options:       []Option{WithTruncatedFrame(true)},
	},
	{
		jfr:           "object-allocation-sample",
		labels:        "",
		expectedCount: 3,
		options:       nil,
	},
	{
		jfr:           "uint64-ref-id",
		labels:        "",
		expectedCount: 5,
		options:       nil,
	},

	{
		jfr:           "parse_failure_repro1",
		labels:        "",
		expectedCount: 1,
		options:       nil,
	},

	{
		jfr:           "wall_tick_sample",
		labels:        "",
		expectedCount: 2,
		options:       nil,
	},
	{
		jfr:           "nativemem",
		labels:        "",
		expectedCount: 1,
		options:       nil,
	},
	{
		jfr:           "new_spancontext",
		labels:        "new_spancontext.labels.gz",
		expectedCount: 1,
		options:       nil,
	},
	{
		jfr:           "wall",
		labels:        "",
		expectedCount: 2,
		options:       nil,
	},
	{
		// Captured with the grafana/async-profiler fork (v4.4.0.0) which emits
		// traceIdHi/traceIdLo; the worker thread set trace id
		// 0123456789abcdeffedcba9876543210 => samples carry a trace_id label.
		jfr:           "traceid",
		labels:        "",
		expectedCount: 1,
		options:       nil,
	},
}

type gprofile struct {
	profile *gpprof.Profile
	proto   *profilev1.Profile
	metric  string
}

func TestDoDump(t *testing.T) {
	assert.False(t, doDump)
}

var parseInput = &ParseInput{
	StartTime:  time.Unix(1706241880, 0),
	EndTime:    time.Unix(1706241890, 0),
	SampleRate: 100,
}

func TestParse(t *testing.T) {
	for _, td := range testFiles {
		t.Run(testName(td), func(t *testing.T) {
			for i, r := range testDataReaders() {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					testOne(t, td, r)
				})
			}
		})
	}
}

func testName(td testdata) string {
	if td.testName != "" {
		return td.testName
	}
	return td.jfr
}

func testOne(t *testing.T, td testdata, r testdataReader) {
	jfrFile := testdataDir + td.jfr + ".jfr.gz"
	jfr, jfrCleanup := r(t, jfrFile)
	ls, lsCleanup := readLabels(t, td, r)
	cleanup := func() {
		jfrCleanup()
		lsCleanup()
	}
	defer cleanup()

	profiles, err := ParseJFR(jfr, parseInput, ls, td.options...)
	require.NoError(t, err)
	cleanup()

	assert.Equal(t, 0, profiles.ParseMetrics.StacktraceNotFound)
	assert.Equal(t, 0, profiles.ParseMetrics.ClassNotFound)
	assert.Equal(t, 0, profiles.ParseMetrics.MethodNotFound)

	gprofiles := toGoogleProfiles(t, profiles.Profiles)
	profiles = nil

	slices.SortFunc(gprofiles, func(i, j gprofile) int {
		return strings.Compare(i.metric, j.metric)
	})
	assert.Equal(t, td.expectedCount, len(gprofiles))

	for i, profile := range gprofiles {
		actual := profileToString(t, profile)
		actualCollapsed := stackCollapseProto(profile.proto, true)
		testFileName := fmt.Sprintf("%s_%d_%s", td.jfr, i, profile.metric)
		if td.testName != "" {
			re := regexp.MustCompile("[^a-zA-Z0-9_]+")
			testFileName = testFileName + "_" + re.ReplaceAllLiteralString(td.testName, "_")
		}
		expectedFile := filepath.Join(testdataDir, fmt.Sprintf("%s_expected.txt.gz", testFileName))
		filePprofDump := filepath.Join(testdataDir, "pprofs", fmt.Sprintf("%s.pprof.gz", testFileName))
		expectedCollapsedFile := filepath.Join(testdataDir, fmt.Sprintf("%s_expected_collapsed.txt.gz", testFileName))
		assert.NotEmpty(t, actual)
		assert.NotEmpty(t, actualCollapsed)
		if doDump {
			profile.proto.TimeNanos = time.Now().UnixNano()
			pprof, err := profile.proto.MarshalVT()
			require.NoError(t, err)
			writeGzipFile(t, filePprofDump, pprof)
			writeGzipFile(t, expectedFile, []byte(actual))
			writeGzipFile(t, expectedCollapsedFile, []byte(actualCollapsed))
		} else {
			expected := readGzipFile(t, expectedFile)
			require.NoError(t, err)
			expectedCollapsed := readGzipFile(t, expectedCollapsedFile)
			require.NoError(t, err)

			assert.Equal(t, string(expected), actual)
			assert.Equal(t, string(expectedCollapsed), actualCollapsed)

			if string(expected) != actual {
				os.WriteFile("actual.txt", []byte(actual), 0644)
				os.WriteFile("expected.txt", expected, 0644)
			}

			if string(expectedCollapsed) != actualCollapsed {
				os.WriteFile("actual_collapsed.txt", []byte(actualCollapsed), 0644)
				os.WriteFile("expected_collapsed.txt", expectedCollapsed, 0644)
			}
		}
	}
}

//todo add tests ingesting parsed testdata into pyroscope container/instance

func profileToString(t *testing.T, profile gprofile) string {
	res := profile.profile.String()
	re := regexp.MustCompile("\nTime: ([^\n]+)\n")
	matches := re.FindAllStringSubmatch(res, -1)
	require.Equal(t, 1, len(matches))
	t2, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", matches[0][1])
	assert.NoError(t, err)
	res = re.ReplaceAllString(res, fmt.Sprintf("\nTime: %d\n", t2.UnixNano()))
	return res
}

func BenchmarkParse(b *testing.B) {
	for _, testfile := range testFiles {
		r := heapReader()
		b.Run(testfile.jfr, func(b *testing.B) {
			jfr, _ := r(b, testdataDir+testfile.jfr+".jfr.gz")
			ls, _ := readLabels(b, testfile, r)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				profiles, err := ParseJFR(jfr, parseInput, ls)
				if err != nil {
					b.Fatalf("Unable to parse JFR file: %s", err)
				}
				if len(profiles.Profiles) == 0 {
					b.Fatalf("No profiles found")
				}
			}
		})
	}
}

// BenchmarkParseAlloc reproduces the allocation spike observed in the
// Pyroscope distributor (distributor-alloc-spike.pprof). The hot paths are:
//   - StackTraceList.Parse: make([]StackFrame, 0, n) per stack trace in each chunk
//   - jfrPprofBuilders.addStacktrace: make([]uint64, 0, nLocs) + make([]int64, ...) per sample
//   - ProfileBuilder.addLocation / addFunction / AddExternalSampleWithLabels: proto heap allocs
//
// Run with -memprofile=mem.pprof to capture an allocation profile for comparison.
func BenchmarkParseAlloc(b *testing.B) {
	r := heapReader()

	cases := []testdata{
		// Largest file: many stack traces → dominates StackTraceList.Parse allocs.
		{jfr: "goland", expectedCount: 5},
		// Multi-chunk variant: exercises chunk-boundary reset of IDMaps.
		{jfr: "goland-multichunk", expectedCount: 5},
		// CPU + lock + alloc: closest to the production event mix that triggered the spike.
		{jfr: "cortex-dev-01__kafka-0__cpu_lock0_alloc0__0", expectedCount: 5},
		// With labels snapshot: exercises AddExternalSampleWithLabels label allocs.
		{jfr: "dump2", labels: "dump2.labels.pb.gz", expectedCount: 4},
	}

	for _, td := range cases {
		b.Run(td.jfr, func(b *testing.B) {
			jfr, _ := r(b, testdataDir+td.jfr+".jfr.gz")
			ls, _ := readLabels(b, td, r)
			b.SetBytes(int64(len(jfr)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				profiles, err := ParseJFR(jfr, parseInput, ls)
				if err != nil {
					b.Fatalf("ParseJFR: %s", err)
				}
				if len(profiles.Profiles) == 0 {
					b.Fatalf("no profiles produced")
				}
			}
		})
	}
}

func toGoogleProfiles(t *testing.T, profiles []Profile) []gprofile {
	res := make([]gprofile, 0, len(profiles))
	for _, profile := range profiles {
		bs, err := profile.Profile.MarshalVT()
		require.NoError(t, err)
		p, err := gpprof.ParseData(bs)
		require.NoError(t, err)

		res = append(res, gprofile{p, profile.Profile, fmt.Sprintf("%s_%s", profile.Metric, sampleTypesToString(p))})
	}
	return res
}

func sampleTypesToString(p *gpprof.Profile) string {
	var sh1 string
	for _, s := range p.SampleType {
		dflt := ""
		sh1 = sh1 + fmt.Sprintf("%s__%s%s ", s.Type, s.Unit, dflt)
	}
	return strings.TrimSpace(sh1)
}

func stackCollapseProto(p *profilev1.Profile, lineNumbers bool) string {
	allZeros := func(a []int64) bool {
		for _, v := range a {
			if v != 0 {
				return false
			}
		}
		return true
	}
	addValues := func(a, b []int64) {
		for i := range a {
			a[i] += b[i]
		}
	}

	type stack struct {
		funcs string
		value []int64
	}
	locMap := make(map[int64]*profilev1.Location)
	funcMap := make(map[int64]*profilev1.Function)
	for _, l := range p.Location {
		locMap[int64(l.Id)] = l
	}
	for _, f := range p.Function {
		funcMap[int64(f.Id)] = f
	}

	var ret []stack
	for _, s := range p.Sample {
		var funcs []string
		for i := range s.LocationId {
			locID := s.LocationId[len(s.LocationId)-1-i]
			loc := locMap[int64(locID)]
			for _, line := range loc.Line {
				f := funcMap[int64(line.FunctionId)]
				fname := p.StringTable[f.Name]
				if lineNumbers {
					fname = fmt.Sprintf("%s:%d", fname, line.Line)
				}
				funcs = append(funcs, fname)
			}
		}

		vv := make([]int64, len(s.Value))
		copy(vv, s.Value)
		ret = append(ret, stack{
			funcs: strings.Join(funcs, ";"),
			value: vv,
		})
	}
	slices.SortFunc(ret, func(i, j stack) int {
		return strings.Compare(i.funcs, j.funcs)
	})
	var unique []stack
	for _, s := range ret {
		if allZeros(s.value) {
			continue
		}
		if len(unique) == 0 {
			unique = append(unique, s)
			continue
		}

		if unique[len(unique)-1].funcs == s.funcs {
			addValues(unique[len(unique)-1].value, s.value)
			continue
		}
		unique = append(unique, s)

	}

	res := make([]string, 0, len(unique))
	for _, s := range unique {
		res = append(res, fmt.Sprintf("%s %v", s.funcs, s.value))
	}
	return strings.Join(res, "\n")
}

func TestProfileId(t *testing.T) {
	assert.Equal(t, "00000000000000ef", profileIdString(0xef))
}

func TestTraceId(t *testing.T) {
	assert.Equal(t, "00000000000000010000000000000002", traceIdString(0x1, 0x2))
	assert.Equal(t, "0123456789abcdeffedcba9876543210", traceIdString(0x0123456789abcdef, 0xfedcba9876543210))
}

// TestTraceIdLabel verifies the trace_id label is emitted from a
// StacktraceCorrelation carrying a nonzero 128-bit trace id, and omitted when
// both halves are zero.
func TestTraceIdLabel(t *testing.T) {
	traceIdOf := func(correlation StacktraceCorrelation) (string, bool) {
		b := NewProfileBuilderWithLabels(0)
		ls := &LabelsSnapshot{Strings: map[int64]string{}}
		b.AddExternalSampleWithLabels([]uint64{1}, []int64{1}, nil, ls, 1, correlation)
		require.Len(t, b.Sample, 1)
		for _, l := range b.Sample[0].Label {
			if b.StringTable[l.Key] == "trace_id" {
				return b.StringTable[l.Str], true
			}
		}
		return "", false
	}

	v, ok := traceIdOf(StacktraceCorrelation{TraceIdHi: 0x1, TraceIdLo: 0x2})
	assert.True(t, ok)
	assert.Equal(t, "00000000000000010000000000000002", v)

	// low half only
	v, ok = traceIdOf(StacktraceCorrelation{TraceIdLo: 0xabc})
	assert.True(t, ok)
	assert.Equal(t, "00000000000000000000000000000abc", v)

	// both zero -> no label
	_, ok = traceIdOf(StacktraceCorrelation{})
	assert.False(t, ok)
}

// TestParseJFROversizedEventSize is a regression test for
// https://github.com/grafana/jfr-parser/issues/90.
// An event whose size field encodes a uint64 with the high bit set causes
// int(size) to be a large negative number, making p.pos = pp + int(size)
// negative. The next p.buf[p.pos] access then panics with
// "index out of range [negative]". The fix validates size against the chunk
// bounds before advancing p.pos.
func TestParseJFROversizedEventSize(t *testing.T) {
	data := readGzipFile(t, testdataDir+"example.jfr.gz")

	// Overwrite the first event's size varint (at offset 68, right after the
	// chunk header) with 9 bytes of 0x80. The modified LEB128 decoder reads
	// the 9th byte with all 8 bits, giving uint64(0x8000000000000000).
	// int(0x8000000000000000) == math.MinInt64, so without the fix
	// p.pos = 68 + MinInt64 which is deeply negative and causes a panic.
	const chunkHeaderSize = 68
	copy(data[chunkHeaderSize:], []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80})

	pi := &ParseInput{
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Minute),
	}

	require.NotPanics(t, func() {
		_, err := ParseJFR(data, pi, nil, WithDisablePanicRecovery(true))
		require.Error(t, err)
	})
}

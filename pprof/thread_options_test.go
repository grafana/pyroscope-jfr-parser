package pprof

import (
	"regexp"
	"strings"
	"testing"

	profilev1 "github.com/grafana/pyroscope/api/gen/proto/go/google/v1"
	"github.com/stretchr/testify/require"
)

// cpuProfile parses the given per-thread testdata file and returns its cpu
// profile, which is where the thread options take effect.
func cpuProfile(t *testing.T, file string, opts ...Option) *profilev1.Profile {
	t.Helper()
	jfr := readGzipFile(t, testdataDir+file+".jfr.gz")
	profiles, err := ParseJFR(jfr, parseInput, nil, opts...)
	require.NoError(t, err)
	for _, p := range profiles.Profiles {
		if p.Metric == "process_cpu" {
			return p.Profile
		}
	}
	t.Fatalf("no process_cpu profile in %s", file)
	return nil
}

func str(p *profilev1.Profile, i int64) string { return p.StringTable[i] }

// funcNames returns the set of function names in the profile.
func funcNames(p *profilev1.Profile) map[string]bool {
	names := map[string]bool{}
	for _, f := range p.Function {
		names[str(p, f.Name)] = true
	}
	return names
}

// leafFunc returns the name of the last location's function in a sample, which
// is the synthetic root frame when WithThreadFrame is enabled.
func leafFunc(p *profilev1.Profile, s *profilev1.Sample) string {
	locID := s.LocationId[len(s.LocationId)-1]
	for _, loc := range p.Location {
		if loc.Id == locID {
			for _, f := range p.Function {
				if f.Id == loc.Line[0].FunctionId {
					return str(p, f.Name)
				}
			}
		}
	}
	return ""
}

func TestWithThreadFrame(t *testing.T) {
	// The "traceid" fixture is recorded per-thread and has a worker thread that
	// burns cpu, so its samples carry a resolvable thread name.
	base := cpuProfile(t, "traceid")
	withFrame := cpuProfile(t, "traceid", WithThreadFrame(true))

	// A thread frame appears as a new root function named after a thread.
	require.False(t, funcNames(base)["worker"], "base profile should not have a thread frame")
	require.True(t, funcNames(withFrame)["worker"], "expected a synthetic frame for the worker thread")

	// Every sample gains exactly one extra (root) location for its thread.
	require.Greater(t, len(withFrame.Sample), 0)
	for i, s := range withFrame.Sample {
		require.NotEmpty(t, leafFunc(withFrame, s), "sample %d has no root frame", i)
	}
}

func TestWithThreadLabel(t *testing.T) {
	// Collapse per-thread names such as "http-nio-auto-1-exec-9" to the pool
	// name "http-nio-auto" so samples aggregate by pool instead of by thread.
	re := regexp.MustCompile(`^(.*?)-\d`)
	pool := func(name string) string {
		if m := re.FindStringSubmatch(name); len(m) > 1 {
			return m[1]
		}
		return ""
	}

	p := cpuProfile(t, "async-profiler", WithThreadLabel("thread_pool", pool))

	labelled := 0
	for _, s := range p.Sample {
		for _, l := range s.Label {
			if str(p, l.Key) == "thread_pool" {
				labelled++
				v := str(p, l.Str)
				require.NotEmpty(t, v)
				// The label carries the collapsed pool name, not a raw
				// "-<n>"-suffixed thread name.
				require.False(t, regexp.MustCompile(`-\d+$`).MatchString(v),
					"thread_pool label %q should be collapsed", v)
			}
		}
	}
	require.Greater(t, labelled, 0, "expected some samples to carry a thread_pool label")

	// A pool label must keep samples on different pools distinct rather than
	// merging them.
	require.True(t, hasLabelValue(p, "thread_pool", "http-nio-auto"),
		"expected the http-nio-auto pool to be labelled")
}

// TestWithThreadLabelNoop verifies the label is opt-in: an empty key or nil fn
// adds nothing.
func TestWithThreadLabelNoop(t *testing.T) {
	p := cpuProfile(t, "async-profiler", WithThreadLabel("", func(string) string { return "x" }))
	for _, s := range p.Sample {
		for _, l := range s.Label {
			require.NotEqual(t, "x", str(p, l.Str))
		}
	}
}

func hasLabelValue(p *profilev1.Profile, key, value string) bool {
	for _, s := range p.Sample {
		for _, l := range s.Label {
			if str(p, l.Key) == key && strings.EqualFold(str(p, l.Str), value) {
				return true
			}
		}
	}
	return false
}

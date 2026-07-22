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
// is the synthetic root frame when the thread frame is enabled.
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

// poolTransform collapses per-thread names such as "http-nio-auto-1-exec-9" to
// the pool name "http-nio-auto".
func poolTransform(name string) string {
	re := regexp.MustCompile(`^(.*?)-\d`)
	if m := re.FindStringSubmatch(name); len(m) > 1 {
		return m[1]
	}
	return ""
}

func TestThreadFrame(t *testing.T) {
	// The "traceid" fixture is recorded per-thread and has a worker thread that
	// burns cpu, so its samples carry a resolvable thread name.
	base := cpuProfile(t, "traceid")
	withFrame := cpuProfile(t, "traceid", WithThreadInfo(ThreadInfoOptions{Frame: true}))

	require.False(t, funcNames(base)["worker"], "base profile should not have a thread frame")
	require.True(t, funcNames(withFrame)["worker"], "expected a synthetic frame for the worker thread")

	require.Greater(t, len(withFrame.Sample), 0)
	for i, s := range withFrame.Sample {
		require.NotEmpty(t, leafFunc(withFrame, s), "sample %d has no root frame", i)
	}
}

// TestThreadFrameTransform checks the shared transform also applies to the
// frame, so the graph can split by pool rather than by individual thread.
func TestThreadFrameTransform(t *testing.T) {
	p := cpuProfile(t, "async-profiler", WithThreadInfo(ThreadInfoOptions{
		Frame:     true,
		Transform: poolTransform,
	}))
	require.True(t, funcNames(p)["http-nio-auto"], "expected a collapsed pool frame")
	require.False(t, funcNames(p)["http-nio-auto-1-exec-9"], "raw thread name should not be a frame")
}

func TestThreadLabel(t *testing.T) {
	p := cpuProfile(t, "async-profiler", WithThreadInfo(ThreadInfoOptions{
		LabelKey:  "thread_pool",
		Transform: poolTransform,
	}))

	labelled := 0
	for _, s := range p.Sample {
		for _, l := range s.Label {
			if str(p, l.Key) == "thread_pool" {
				labelled++
				v := str(p, l.Str)
				require.NotEmpty(t, v)
				require.False(t, regexp.MustCompile(`-\d+$`).MatchString(v),
					"thread_pool label %q should be collapsed", v)
			}
		}
	}
	require.Greater(t, labelled, 0, "expected some samples to carry a thread_pool label")
	require.True(t, hasLabelValue(p, "thread_pool", "http-nio-auto"),
		"expected the http-nio-auto pool to be labelled")
}

// TestThreadInfoNoop verifies it is opt-in: with neither frame nor label key,
// nothing is added even when a transform is supplied.
func TestThreadInfoNoop(t *testing.T) {
	p := cpuProfile(t, "async-profiler", WithThreadInfo(ThreadInfoOptions{
		Transform: func(string) string { return "x" },
	}))
	require.False(t, funcNames(p)["x"])
	for _, s := range p.Sample {
		for _, l := range s.Label {
			require.NotEqual(t, "x", str(p, l.Str))
		}
	}
}

// TestThreadTransformMemoized checks the transform runs once per distinct thread
// name, not once per sample.
func TestThreadTransformMemoized(t *testing.T) {
	seen := map[string]int{}
	p := cpuProfile(t, "async-profiler", WithThreadInfo(ThreadInfoOptions{
		LabelKey: "thread_pool",
		Transform: func(name string) string {
			seen[name]++
			return poolTransform(name)
		},
	}))
	require.Greater(t, len(p.Sample), len(seen),
		"expected more samples than distinct transform inputs")
	for name, n := range seen {
		require.Equal(t, 1, n, "transform called %d times for %q, expected once", n, name)
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

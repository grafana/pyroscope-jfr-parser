package pprof

import (
	"fmt"
	"io"

	"github.com/grafana/jfr-parser/parser"
)

type pprofOptions struct {
	truncatedFrame       bool
	disablePanicRecovery bool
	threadFrame          bool
	threadLabelKey       string
	threadLabelFn        func(threadName string) string
}
type Option func(*pprofOptions)

func WithTruncatedFrame(v bool) Option {
	return func(o *pprofOptions) {
		o.truncatedFrame = v
	}
}

func WithDisablePanicRecovery(v bool) Option {
	return func(o *pprofOptions) {
		o.disablePanicRecovery = v
	}
}

// WithThreadFrame, when enabled, adds the sampled thread's name as a synthetic
// root frame beneath each execution-sample stack, so flame graphs split by the
// thread that ran the sample. Requires the profile to have been recorded with
// per-thread sampling. Off by default.
func WithThreadFrame(v bool) Option {
	return func(o *pprofOptions) {
		o.threadFrame = v
	}
}

// WithThreadLabel adds a pprof sample label to each execution sample whose value
// is derived from the sampled thread's name by fn. The caller supplies the
// naming convention (e.g. extracting a thread-pool name via regexp), keeping
// this package free of application-specific parsing. Requires per-thread
// sampling. If fn returns "" for a sample, no label is added for it. A no-op
// unless key is non-empty and fn is non-nil.
func WithThreadLabel(key string, fn func(threadName string) string) Option {
	return func(o *pprofOptions) {
		o.threadLabelKey = key
		o.threadLabelFn = fn
	}
}

func ParseJFR(body []byte, pi *ParseInput, jfrLabels *LabelsSnapshot, opts ...Option) (res *Profiles, err error) {
	o := &pprofOptions{
		truncatedFrame:       false,
		disablePanicRecovery: false,
	}
	for i := range opts {
		opts[i](o)
	}

	if !o.disablePanicRecovery {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("jfr parser panic: %v", r)
			}
		}()
	}

	p := parser.NewParser(body, parser.Options{
		SymbolProcessor: parser.ProcessSymbols,
	})
	return parse(p, pi, jfrLabels, o)
}

func parse(parser *parser.Parser, piOriginal *ParseInput, jfrLabels *LabelsSnapshot, opt *pprofOptions) (result *Profiles, err error) {
	var event string

	builders := newJfrPprofBuilders(parser, jfrLabels, piOriginal, opt)

	var values = [2]int64{1, 0}

	for {
		typ, err := parser.ParseEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("jfr parser ParseEvent error: %w", err)
		}
		switch typ {
		case parser.TypeMap.T_EXECUTION_SAMPLE:
			ts := parser.GetThreadState(parser.ExecutionSample.State)
			correlation := StacktraceCorrelation{
				ContextId: parser.ExecutionSample.ContextId,
				SpanId:    parser.ExecutionSample.SpanId,
				SpanName:  parser.ExecutionSample.SpanName,
				TraceIdHi: parser.ExecutionSample.TraceIdHi,
				TraceIdLo: parser.ExecutionSample.TraceIdLo,
			}
			// When async-profiler runs with -t, each sample carries the thread it
			// ran on. Managed JVM threads carry their full name in JavaName;
			// native threads (GC, JIT compiler, VM tasks) have no JavaName and
			// are identified only by OsName, so fall back to it. The name is used
			// only when a thread frame or a thread label was requested.
			if opt.threadFrame || (opt.threadLabelKey != "" && opt.threadLabelFn != nil) {
				if t := parser.GetThread(parser.ExecutionSample.SampledThread); t != nil {
					name := t.JavaName
					if name == "" {
						name = t.OsName
					}
					if opt.threadFrame {
						correlation.ThreadName = name
					}
					if name != "" && opt.threadLabelFn != nil && opt.threadLabelKey != "" {
						if v := opt.threadLabelFn(name); v != "" {
							correlation.ThreadLabelKey = opt.threadLabelKey
							correlation.ThreadLabelValue = v
						}
					}
				}
			}
			if ts != nil && ts.Name != "STATE_SLEEPING" {
				builders.addStacktrace(sampleTypeCPU, correlation, parser.ExecutionSample.StackTrace, values[:1])
			}
			if event == "wall" {
				builders.addStacktrace(sampleTypeWall, correlation, parser.ExecutionSample.StackTrace, values[:1])
			}
		case parser.TypeMap.T_WALL_CLOCK_SAMPLE:
			values[0] = int64(parser.WallClockSample.Samples)
			correlation := StacktraceCorrelation{
				ContextId: parser.WallClockSample.ContextId,
				SpanId:    parser.WallClockSample.SpanId,
				SpanName:  parser.WallClockSample.SpanName,
				TraceIdHi: parser.WallClockSample.TraceIdHi,
				TraceIdLo: parser.WallClockSample.TraceIdLo,
			}
			ts := parser.GetThreadState(parser.WallClockSample.State)
			if ts != nil && ts.Name == "STATE_RUNNABLE" && event == "wall" {
				builders.addStacktrace(sampleTypeCPU, correlation, parser.WallClockSample.StackTrace, values[:1])
			}
			builders.addStacktrace(sampleTypeWall, correlation, parser.WallClockSample.StackTrace, values[:1])
		case parser.TypeMap.T_ALLOC_IN_NEW_TLAB:
			values[1] = int64(parser.ObjectAllocationInNewTLAB.TlabSize)
			correlation := StacktraceCorrelation{
				ContextId: parser.ObjectAllocationInNewTLAB.ContextId,
				SpanId:    parser.ObjectAllocationInNewTLAB.SpanId,
				SpanName:  parser.ObjectAllocationInNewTLAB.SpanName,
				TraceIdHi: parser.ObjectAllocationInNewTLAB.TraceIdHi,
				TraceIdLo: parser.ObjectAllocationInNewTLAB.TraceIdLo,
			}
			builders.addStacktrace(sampleTypeInTLAB, correlation, parser.ObjectAllocationInNewTLAB.StackTrace, values[:2])
		case parser.TypeMap.T_ALLOC_OUTSIDE_TLAB:
			values[1] = int64(parser.ObjectAllocationOutsideTLAB.AllocationSize)
			correlation := StacktraceCorrelation{
				ContextId: parser.ObjectAllocationOutsideTLAB.ContextId,
				SpanId:    parser.ObjectAllocationOutsideTLAB.SpanId,
				SpanName:  parser.ObjectAllocationOutsideTLAB.SpanName,
				TraceIdHi: parser.ObjectAllocationOutsideTLAB.TraceIdHi,
				TraceIdLo: parser.ObjectAllocationOutsideTLAB.TraceIdLo,
			}
			builders.addStacktrace(sampleTypeOutTLAB, correlation, parser.ObjectAllocationOutsideTLAB.StackTrace, values[:2])
		case parser.TypeMap.T_ALLOC_SAMPLE:
			values[1] = int64(parser.ObjectAllocationSample.Weight)
			builders.addStacktrace(sampleTypeAllocSample, StacktraceCorrelation{}, parser.ObjectAllocationSample.StackTrace, values[:2])
		case parser.TypeMap.T_MONITOR_ENTER:
			values[1] = int64(parser.JavaMonitorEnter.Duration)
			correlation := StacktraceCorrelation{
				ContextId: parser.JavaMonitorEnter.ContextId,
				SpanId:    parser.JavaMonitorEnter.SpanId,
				SpanName:  parser.JavaMonitorEnter.SpanName,
				TraceIdHi: parser.JavaMonitorEnter.TraceIdHi,
				TraceIdLo: parser.JavaMonitorEnter.TraceIdLo,
			}
			builders.addStacktrace(sampleTypeLock, correlation, parser.JavaMonitorEnter.StackTrace, values[:2])
		case parser.TypeMap.T_THREAD_PARK:
			values[1] = int64(parser.ThreadPark.Duration)
			builders.addStacktrace(sampleTypeThreadPark, StacktraceCorrelation{}, parser.ThreadPark.StackTrace, values[:2])
		case parser.TypeMap.T_LIVE_OBJECT:
			builders.addStacktrace(sampleTypeLiveObject, StacktraceCorrelation{}, parser.LiveObject.StackTrace, values[:1])
		case parser.TypeMap.T_MALLOC:
			values[1] = int64(parser.Malloc.Size)
			builders.addStacktrace(sampleTypeMalloc, StacktraceCorrelation{}, parser.Malloc.StackTrace, values[:2])
		case parser.TypeMap.T_ACTIVE_SETTING:
			if parser.ActiveSetting.Name == "event" {
				event = parser.ActiveSetting.Value
			}
		}
	}

	result = builders.build(event)

	return result, nil
}

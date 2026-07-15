package pprof

import (
	"time"

	profilev1 "github.com/grafana/pyroscope/api/gen/proto/go/google/v1"
	typesv1 "github.com/grafana/pyroscope/api/gen/proto/go/types/v1"
)

type Labels []*typesv1.LabelPair

type ParseInput struct {
	StartTime  time.Time
	EndTime    time.Time
	SampleRate int64
}

type Profiles struct {
	Profiles []Profile
	JFREvent string

	// ProcessRuntimeName is read from java.vm.name because JFR InitialSystemProperty
	// records expose the VM property set, not the full java.lang.System properties.
	ProcessRuntimeName string
	// ProcessRuntimeVersion is read from java.vm.version because JFR InitialSystemProperty
	// records expose the VM property set, not the full java.lang.System properties.
	ProcessRuntimeVersion string
	// ProcessRuntimeVersionMajor is derived from java.vm.specification.version.
	// Legacy versions like 1.8 are normalized to 8.
	ProcessRuntimeVersionMajor string

	ParseMetrics ParseMetrics
}

type Profile struct {
	Profile *profilev1.Profile
	Metric  string
}

type ParseMetrics struct {
	StacktraceNotFound int
	ClassNotFound      int
	MethodNotFound     int
}

package jfrwriter_test

import (
	"io"
	"testing"

	"github.com/grafana/jfr-parser/jfrwriter"
	"github.com/grafana/jfr-parser/parser"
	"github.com/grafana/jfr-parser/parser/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// addMinimalCPool adds the minimal set of constant pool entries required
// for the parser to process events with stack traces and threads.
func addMinimalCPool(b *jfrwriter.Builder) {
	// Symbols
	b.AddCPoolEntry(jfrwriter.T_SYMBOL, 1, jfrwriter.SerializeSymbol("main"))
	b.AddCPoolEntry(jfrwriter.T_SYMBOL, 2, jfrwriter.SerializeSymbol("()V"))
	b.AddCPoolEntry(jfrwriter.T_SYMBOL, 3, jfrwriter.SerializeSymbol("com/example"))
	b.AddCPoolEntry(jfrwriter.T_SYMBOL, 4, jfrwriter.SerializeSymbol("MyClass"))

	// Frame type and thread state
	b.AddCPoolEntry(jfrwriter.T_FRAME_TYPE, 1, jfrwriter.SerializeFrameType("Interpreted"))
	b.AddCPoolEntry(jfrwriter.T_THREAD_STATE, 1, jfrwriter.SerializeThreadState("STATE_RUNNABLE"))

	// Log level (required by async-profiler metadata)
	b.AddCPoolEntry(jfrwriter.T_LOG_LEVEL, 1, jfrwriter.SerializeLogLevel("INFO"))

	// Class loader, package, class
	b.AddCPoolEntry(jfrwriter.T_CLASS_LOADER, 1, jfrwriter.SerializeClassLoader(0, 0))
	b.AddCPoolEntry(jfrwriter.T_PACKAGE, 1, jfrwriter.SerializePackage(3))
	b.AddCPoolEntry(jfrwriter.T_CLASS, 1, jfrwriter.SerializeClass(1, 4, 1, 0))

	// Method
	b.AddCPoolEntry(jfrwriter.T_METHOD, 1, jfrwriter.SerializeMethod(1, 1, 2, 0, false))

	// Thread
	b.AddCPoolEntry(jfrwriter.T_THREAD, 1, jfrwriter.SerializeThread("main", 1, "main-thread", 1))

	// Stack trace with one frame
	b.AddCPoolEntry(jfrwriter.T_STACK_TRACE, 1, jfrwriter.SerializeStackTrace(false, [][]byte{
		jfrwriter.SerializeStackFrame(1, 42, 0, 1),
	}))
}

func TestRoundTripExecutionSample(t *testing.T) {
	b := jfrwriter.NewBuilder()
	b.AddClasses(jfrwriter.AllClasses()...)
	addMinimalCPool(b)

	b.AddEvent(jfrwriter.T_EXECUTION_SAMPLE,
		jfrwriter.SerializeExecutionSample(500, 1, 1, 1))

	data, err := b.Build()
	require.NoError(t, err)

	p := parser.NewParser(data, parser.Options{})
	typ, err := p.ParseEvent()
	require.NoError(t, err)
	assert.Equal(t, p.TypeMap.T_EXECUTION_SAMPLE, typ)
	assert.Equal(t, types.ThreadRef(1), p.ExecutionSample.SampledThread)
	assert.Equal(t, types.StackTraceRef(1), p.ExecutionSample.StackTrace)
	assert.Equal(t, types.ThreadStateRef(1), p.ExecutionSample.State)

	_, err = p.ParseEvent()
	assert.Equal(t, io.EOF, err)
}

// TestRoundTripObjectAllocationSample exercises the T_ALLOC_SAMPLE code path
// in parser.go:199-208. This is the regression test for issue #89:
// https://github.com/grafana/jfr-parser/issues/89
//
// The bug is a missing "continue" after the skip in the T_ALLOC_SAMPLE case.
// This test generates a valid JFR with ObjectAllocationSample events and verifies
// the parser handles them correctly.
func TestRoundTripObjectAllocationSample(t *testing.T) {
	b := jfrwriter.NewBuilder()
	b.AddClasses(jfrwriter.AllClasses()...)
	addMinimalCPool(b)

	b.AddEvent(jfrwriter.T_ALLOC_SAMPLE,
		jfrwriter.SerializeObjectAllocationSample(500, 1, 1, 1, 1024))

	data, err := b.Build()
	require.NoError(t, err)

	p := parser.NewParser(data, parser.Options{})
	typ, err := p.ParseEvent()
	require.NoError(t, err)
	assert.Equal(t, p.TypeMap.T_ALLOC_SAMPLE, typ)
	assert.Equal(t, uint64(1024), p.ObjectAllocationSample.Weight)
	assert.Equal(t, types.ThreadRef(1), p.ObjectAllocationSample.EventThread)
	assert.Equal(t, types.StackTraceRef(1), p.ObjectAllocationSample.StackTrace)
	assert.Equal(t, types.ClassRef(1), p.ObjectAllocationSample.ObjectClass)

	_, err = p.ParseEvent()
	assert.Equal(t, io.EOF, err)
}

func TestRoundTripMultipleEvents(t *testing.T) {
	b := jfrwriter.NewBuilder()
	b.AddClasses(jfrwriter.AllClasses()...)
	addMinimalCPool(b)

	// Add two execution samples and one alloc sample
	b.AddEvent(jfrwriter.T_EXECUTION_SAMPLE,
		jfrwriter.SerializeExecutionSample(100, 1, 1, 1))
	b.AddEvent(jfrwriter.T_ALLOC_SAMPLE,
		jfrwriter.SerializeObjectAllocationSample(200, 1, 1, 1, 2048))
	b.AddEvent(jfrwriter.T_EXECUTION_SAMPLE,
		jfrwriter.SerializeExecutionSample(300, 1, 1, 1))

	data, err := b.Build()
	require.NoError(t, err)

	p := parser.NewParser(data, parser.Options{})

	// First event: execution sample
	typ, err := p.ParseEvent()
	require.NoError(t, err)
	assert.Equal(t, p.TypeMap.T_EXECUTION_SAMPLE, typ)
	assert.Equal(t, uint64(100), p.ExecutionSample.StartTime)

	// Second event: alloc sample
	typ, err = p.ParseEvent()
	require.NoError(t, err)
	assert.Equal(t, p.TypeMap.T_ALLOC_SAMPLE, typ)
	assert.Equal(t, uint64(2048), p.ObjectAllocationSample.Weight)

	// Third event: execution sample
	typ, err = p.ParseEvent()
	require.NoError(t, err)
	assert.Equal(t, p.TypeMap.T_EXECUTION_SAMPLE, typ)
	assert.Equal(t, uint64(300), p.ExecutionSample.StartTime)

	_, err = p.ParseEvent()
	assert.Equal(t, io.EOF, err)
}

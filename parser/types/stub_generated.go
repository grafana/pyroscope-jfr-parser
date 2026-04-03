//go:build jfrparserstub

package types

// Stubs for generated types. Used to bootstrap the code generator
// when generated files are absent. Build with -tags jfrparserstub.

type MethodRef uint64
type FrameTypeRef uint64
type ClassRef uint64
type BindClass struct{}
type BindThread struct{}
type BindFrameType struct{}
type BindThreadState struct{}
type BindStackTrace struct{}
type BindMethod struct{}
type BindPackage struct{}
type BindSymbol struct{}
type BindLogLevel struct{}
type BindStackFrame struct{}
type BindClassLoader struct{}
type BindExecutionSample struct{}
type BindWallClockSample struct{}
type BindObjectAllocationInNewTLAB struct{}
type BindObjectAllocationOutsideTLAB struct{}
type BindObjectAllocationSample struct{}
type BindLiveObject struct{}
type BindJavaMonitorEnter struct{}
type BindThreadPark struct{}
type BindActiveSetting struct{}
type BindMalloc struct{}
type BindFree struct{}
type BindString struct{}
type BindSkipConstantPool struct{}

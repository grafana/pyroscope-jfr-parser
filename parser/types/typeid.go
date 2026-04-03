package types

import (
	"fmt"

	"golang.org/x/text/encoding"
)

type TypeID int64

const UnsetTypeID TypeID = -1

type BindingResolver interface {
	Resolve(typeMap *TypeMap) error
}

type TypeBinding[B any] struct {
	Name     string
	TypeID   TypeID
	Bind     *B
	Factory  func(*MetadataClass, *TypeMap) *B
	Required bool
}

func (tb *TypeBinding[B]) Resolve(typeMap *TypeMap) error {
	cls := typeMap.NameMap[tb.Name]
	if cls != nil {
		tb.TypeID = cls.ID
		tb.Bind = tb.Factory(cls, typeMap)
	} else if tb.Required {
		return fmt.Errorf("missing %q", tb.Name)
	} else {
		tb.TypeID = UnsetTypeID
		tb.Bind = nil
	}
	return nil
}

type TypeMap struct {
	IDMap   map[TypeID]*MetadataClass
	NameMap map[string]*MetadataClass

	T_STRING  TypeBinding[BindString]
	T_INT     TypeID
	T_LONG    TypeID
	T_SHORT   TypeID
	T_FLOAT   TypeID
	T_BOOLEAN TypeID

	T_CLASS        TypeBinding[BindClass]
	T_THREAD       TypeBinding[BindThread]
	T_FRAME_TYPE   TypeBinding[BindFrameType]
	T_THREAD_STATE TypeBinding[BindThreadState]
	T_STACK_TRACE  TypeBinding[BindStackTrace]
	T_METHOD       TypeBinding[BindMethod]
	T_PACKAGE      TypeBinding[BindPackage]
	T_SYMBOL       TypeBinding[BindSymbol]
	T_LOG_LEVEL    TypeBinding[BindLogLevel]

	T_STACK_FRAME  TypeBinding[BindStackFrame]
	T_CLASS_LOADER TypeBinding[BindClassLoader]

	T_EXECUTION_SAMPLE   TypeBinding[BindExecutionSample]
	T_WALL_CLOCK_SAMPLE  TypeBinding[BindWallClockSample]
	T_ALLOC_IN_NEW_TLAB  TypeBinding[BindObjectAllocationInNewTLAB]
	T_ALLOC_OUTSIDE_TLAB TypeBinding[BindObjectAllocationOutsideTLAB]
	T_ALLOC_SAMPLE       TypeBinding[BindObjectAllocationSample]
	T_LIVE_OBJECT        TypeBinding[BindLiveObject]
	T_MONITOR_ENTER      TypeBinding[BindJavaMonitorEnter]
	T_THREAD_PARK        TypeBinding[BindThreadPark]
	T_ACTIVE_SETTING     TypeBinding[BindActiveSetting]
	T_MALLOC             TypeBinding[BindMalloc]
	T_FREE               TypeBinding[BindFree]

	ISO8859_1Decoder *encoding.Decoder

	bindings []BindingResolver
}

func addBinding[B any](tm *TypeMap, tb *TypeBinding[B], name string, factory func(*MetadataClass, *TypeMap) *B, required bool) {
	tb.Name = name
	tb.Factory = factory
	tb.Required = required
	tm.bindings = append(tm.bindings, tb)
}

// initBindings registers all type bindings in dependency order.
// Must be called on the final TypeMap location (not a temporary)
// since it stores pointers to the TypeMap's fields.
func (tm *TypeMap) InitBindings() {
	tm.bindings = tm.bindings[:0]
	// Order matters: factories check typeMap.T_*.TypeID of dependencies,
	// so dependencies must resolve before dependents.
	// Required cpool types (topologically sorted)
	addBinding(tm, &tm.T_STRING, "java.lang.String", NewBindString, true)
	addBinding(tm, &tm.T_SYMBOL, "jdk.types.Symbol", NewBindSymbol, true)
	addBinding(tm, &tm.T_PACKAGE, "jdk.types.Package", NewBindPackage, true)
	addBinding(tm, &tm.T_FRAME_TYPE, "jdk.types.FrameType", NewBindFrameType, true)
	addBinding(tm, &tm.T_THREAD_STATE, "jdk.types.ThreadState", NewBindThreadState, true)
	addBinding(tm, &tm.T_THREAD, "java.lang.Thread", NewBindThread, true)
	addBinding(tm, &tm.T_CLASS_LOADER, "jdk.types.ClassLoader", NewBindClassLoader, true)
	addBinding(tm, &tm.T_CLASS, "java.lang.Class", NewBindClass, true)
	addBinding(tm, &tm.T_METHOD, "jdk.types.Method", NewBindMethod, true)
	addBinding(tm, &tm.T_STACK_FRAME, "jdk.types.StackFrame", NewBindStackFrame, true)
	addBinding(tm, &tm.T_STACK_TRACE, "jdk.types.StackTrace", NewBindStackTrace, true)
	// Optional cpool type
	addBinding(tm, &tm.T_LOG_LEVEL, "profiler.types.LogLevel", NewBindLogLevel, false)
	// Optional event types
	addBinding(tm, &tm.T_EXECUTION_SAMPLE, "jdk.ExecutionSample", NewBindExecutionSample, false)
	addBinding(tm, &tm.T_WALL_CLOCK_SAMPLE, "profiler.WallClockSample", NewBindWallClockSample, false)
	addBinding(tm, &tm.T_MALLOC, "profiler.Malloc", NewBindMalloc, false)
	addBinding(tm, &tm.T_FREE, "profiler.Free", NewBindFree, false)
	addBinding(tm, &tm.T_ALLOC_IN_NEW_TLAB, "jdk.ObjectAllocationInNewTLAB", NewBindObjectAllocationInNewTLAB, false)
	addBinding(tm, &tm.T_ALLOC_OUTSIDE_TLAB, "jdk.ObjectAllocationOutsideTLAB", NewBindObjectAllocationOutsideTLAB, false)
	addBinding(tm, &tm.T_ALLOC_SAMPLE, "jdk.ObjectAllocationSample", NewBindObjectAllocationSample, false)
	addBinding(tm, &tm.T_MONITOR_ENTER, "jdk.JavaMonitorEnter", NewBindJavaMonitorEnter, false)
	addBinding(tm, &tm.T_THREAD_PARK, "jdk.ThreadPark", NewBindThreadPark, false)
	addBinding(tm, &tm.T_LIVE_OBJECT, "profiler.LiveObject", NewBindLiveObject, false)
	addBinding(tm, &tm.T_ACTIVE_SETTING, "jdk.ActiveSetting", NewBindActiveSetting, false)
}

func (tm *TypeMap) Bindings() []BindingResolver {
	return tm.bindings
}

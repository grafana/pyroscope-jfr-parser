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
}

func NewTypeMap() TypeMap {
	return TypeMap{
		// Required cpool types
		T_STRING:       TypeBinding[BindString]{Name: "java.lang.String", Factory: NewBindString, Required: true},
		T_FRAME_TYPE:   TypeBinding[BindFrameType]{Name: "jdk.types.FrameType", Factory: NewBindFrameType, Required: true},
		T_THREAD_STATE: TypeBinding[BindThreadState]{Name: "jdk.types.ThreadState", Factory: NewBindThreadState, Required: true},
		T_THREAD:       TypeBinding[BindThread]{Name: "java.lang.Thread", Factory: NewBindThread, Required: true},
		T_CLASS:        TypeBinding[BindClass]{Name: "java.lang.Class", Factory: NewBindClass, Required: true},
		T_METHOD:       TypeBinding[BindMethod]{Name: "jdk.types.Method", Factory: NewBindMethod, Required: true},
		T_PACKAGE:      TypeBinding[BindPackage]{Name: "jdk.types.Package", Factory: NewBindPackage, Required: true},
		T_SYMBOL:       TypeBinding[BindSymbol]{Name: "jdk.types.Symbol", Factory: NewBindSymbol, Required: true},
		T_STACK_TRACE:  TypeBinding[BindStackTrace]{Name: "jdk.types.StackTrace", Factory: NewBindStackTrace, Required: true},
		T_STACK_FRAME:  TypeBinding[BindStackFrame]{Name: "jdk.types.StackFrame", Factory: NewBindStackFrame, Required: true},
		T_CLASS_LOADER: TypeBinding[BindClassLoader]{Name: "jdk.types.ClassLoader", Factory: NewBindClassLoader, Required: true},
		// Optional cpool type
		T_LOG_LEVEL: TypeBinding[BindLogLevel]{Name: "profiler.types.LogLevel", Factory: NewBindLogLevel},
		// Optional event types
		T_EXECUTION_SAMPLE:   TypeBinding[BindExecutionSample]{Name: "jdk.ExecutionSample", Factory: NewBindExecutionSample},
		T_WALL_CLOCK_SAMPLE:  TypeBinding[BindWallClockSample]{Name: "profiler.WallClockSample", Factory: NewBindWallClockSample},
		T_MALLOC:             TypeBinding[BindMalloc]{Name: "profiler.Malloc", Factory: NewBindMalloc},
		T_FREE:               TypeBinding[BindFree]{Name: "profiler.Free", Factory: NewBindFree},
		T_ALLOC_IN_NEW_TLAB:  TypeBinding[BindObjectAllocationInNewTLAB]{Name: "jdk.ObjectAllocationInNewTLAB", Factory: NewBindObjectAllocationInNewTLAB},
		T_ALLOC_OUTSIDE_TLAB: TypeBinding[BindObjectAllocationOutsideTLAB]{Name: "jdk.ObjectAllocationOutsideTLAB", Factory: NewBindObjectAllocationOutsideTLAB},
		T_ALLOC_SAMPLE:       TypeBinding[BindObjectAllocationSample]{Name: "jdk.ObjectAllocationSample", Factory: NewBindObjectAllocationSample},
		T_MONITOR_ENTER:      TypeBinding[BindJavaMonitorEnter]{Name: "jdk.JavaMonitorEnter", Factory: NewBindJavaMonitorEnter},
		T_THREAD_PARK:        TypeBinding[BindThreadPark]{Name: "jdk.ThreadPark", Factory: NewBindThreadPark},
		T_LIVE_OBJECT:        TypeBinding[BindLiveObject]{Name: "profiler.LiveObject", Factory: NewBindLiveObject},
		T_ACTIVE_SETTING:     TypeBinding[BindActiveSetting]{Name: "jdk.ActiveSetting", Factory: NewBindActiveSetting},
	}
}

func (tm *TypeMap) Bindings() []BindingResolver {
	// Order matters: factories check typeMap.T_*.TypeID of dependencies,
	// so dependencies must resolve before dependents.
	// String has no deps. Symbol depends on String. Package depends on Symbol.
	// ClassLoader depends on Class+Symbol. Class depends on Symbol+ClassLoader+Package.
	// Method depends on Class+Symbol. FrameType depends on String.
	// ThreadState depends on String. Thread depends on String.
	// StackFrame depends on Method+FrameType. StackTrace depends on StackFrame.
	// LogLevel depends on String.
	return []BindingResolver{
		&tm.T_STRING,
		&tm.T_SYMBOL,
		&tm.T_PACKAGE,
		&tm.T_FRAME_TYPE,
		&tm.T_THREAD_STATE,
		&tm.T_LOG_LEVEL,
		&tm.T_THREAD,
		&tm.T_CLASS_LOADER,
		&tm.T_CLASS,
		&tm.T_METHOD,
		&tm.T_STACK_FRAME,
		&tm.T_STACK_TRACE,
		&tm.T_EXECUTION_SAMPLE,
		&tm.T_WALL_CLOCK_SAMPLE,
		&tm.T_ALLOC_IN_NEW_TLAB,
		&tm.T_ALLOC_OUTSIDE_TLAB,
		&tm.T_ALLOC_SAMPLE,
		&tm.T_LIVE_OBJECT,
		&tm.T_MONITOR_ENTER,
		&tm.T_THREAD_PARK,
		&tm.T_ACTIVE_SETTING,
		&tm.T_MALLOC,
		&tm.T_FREE,
	}
}

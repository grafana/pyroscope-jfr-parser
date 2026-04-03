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

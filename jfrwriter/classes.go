package jfrwriter

import "github.com/grafana/jfr-parser/parser/types/def"

// Predefined type IDs matching async-profiler conventions.
// These mirror the definitions in internal/cmd/gen/types.go.
var (
	T_METADATA                = def.TypeID(0)
	T_CPOOL                   = def.TypeID(1)
	T_BOOLEAN                 = def.TypeID(4)
	T_CHAR                    = def.TypeID(5)
	T_FLOAT                   = def.TypeID(6)
	T_DOUBLE                  = def.TypeID(7)
	T_BYTE                    = def.TypeID(8)
	T_SHORT                   = def.TypeID(9)
	T_INT                     = def.TypeID(10)
	T_LONG                    = def.TypeID(11)
	T_STRING                  = def.TypeID(20)
	T_CLASS                   = def.TypeID(21)
	T_THREAD                  = def.TypeID(22)
	T_CLASS_LOADER            = def.TypeID(23)
	T_FRAME_TYPE              = def.TypeID(24)
	T_THREAD_STATE            = def.TypeID(25)
	T_STACK_TRACE             = def.TypeID(26)
	T_STACK_FRAME             = def.TypeID(27)
	T_METHOD                  = def.TypeID(28)
	T_PACKAGE                 = def.TypeID(29)
	T_SYMBOL                  = def.TypeID(30)
	T_LOG_LEVEL               = def.TypeID(31)
	T_EVENT                   = def.TypeID(100)
	T_EXECUTION_SAMPLE        = def.TypeID(101)
	T_ALLOC_IN_NEW_TLAB       = def.TypeID(102)
	T_ALLOC_OUTSIDE_TLAB      = def.TypeID(103)
	T_MONITOR_ENTER           = def.TypeID(104)
	T_THREAD_PARK             = def.TypeID(105)
	T_CPU_LOAD                = def.TypeID(106)
	T_ACTIVE_RECORDING        = def.TypeID(107)
	T_ACTIVE_SETTING          = def.TypeID(108)
	T_OS_INFORMATION          = def.TypeID(109)
	T_CPU_INFORMATION         = def.TypeID(110)
	T_JVM_INFORMATION         = def.TypeID(111)
	T_INITIAL_SYSTEM_PROPERTY = def.TypeID(112)
	T_NATIVE_LIBRARY          = def.TypeID(113)
	T_LOG                     = def.TypeID(114)
	T_LIVE_OBJECT             = def.TypeID(115)
	T_WALL_CLOCK_SAMPLE       = def.TypeID(118)
	T_MALLOC                  = def.TypeID(119)
	T_FREE                    = def.TypeID(120)
	T_ANNOTATION              = def.TypeID(200)
	T_LABEL                   = def.TypeID(201)
	T_CATEGORY                = def.TypeID(202)
	T_TIMESTAMP               = def.TypeID(203)
	T_TIMESPAN                = def.TypeID(204)
	T_DATA_AMOUNT             = def.TypeID(205)
	T_MEMORY_ADDRESS          = def.TypeID(206)
	T_UNSIGNED                = def.TypeID(207)
	T_PERCENTAGE              = def.TypeID(208)
	T_ALLOC_SAMPLE            = def.TypeID(209)
)

// AllClasses returns all predefined class definitions used by async-profiler.
func AllClasses() []def.Class {
	return []def.Class{
		// Primitives
		{Name: "boolean", ID: T_BOOLEAN},
		{Name: "char", ID: T_CHAR},
		{Name: "float", ID: T_FLOAT},
		{Name: "double", ID: T_DOUBLE},
		{Name: "byte", ID: T_BYTE},
		{Name: "short", ID: T_SHORT},
		{Name: "int", ID: T_INT},
		{Name: "long", ID: T_LONG},
		{Name: "java.lang.String", ID: T_STRING},

		// Constant pool types
		{Name: "java.lang.Class", ID: T_CLASS, Fields: []def.Field{
			{Name: "classLoader", Type: T_CLASS_LOADER, ConstantPool: true},
			{Name: "name", Type: T_SYMBOL, ConstantPool: true},
			{Name: "package", Type: T_PACKAGE, ConstantPool: true},
			{Name: "modifiers", Type: T_INT},
		}},
		{Name: "java.lang.Thread", ID: T_THREAD, Fields: []def.Field{
			{Name: "osName", Type: T_STRING},
			{Name: "osThreadId", Type: T_LONG},
			{Name: "javaName", Type: T_STRING},
			{Name: "javaThreadId", Type: T_LONG},
		}},
		{Name: "jdk.types.ClassLoader", ID: T_CLASS_LOADER, Fields: []def.Field{
			{Name: "type", Type: T_CLASS, ConstantPool: true},
			{Name: "name", Type: T_SYMBOL, ConstantPool: true},
		}},
		{Name: "jdk.types.FrameType", ID: T_FRAME_TYPE, Fields: []def.Field{
			{Name: "description", Type: T_STRING},
		}},
		{Name: "jdk.types.ThreadState", ID: T_THREAD_STATE, Fields: []def.Field{
			{Name: "name", Type: T_STRING},
		}},
		{Name: "jdk.types.StackTrace", ID: T_STACK_TRACE, Fields: []def.Field{
			{Name: "truncated", Type: T_BOOLEAN},
			{Name: "frames", Type: T_STACK_FRAME, Array: true},
		}},
		{Name: "jdk.types.StackFrame", ID: T_STACK_FRAME, Fields: []def.Field{
			{Name: "method", Type: T_METHOD, ConstantPool: true},
			{Name: "lineNumber", Type: T_INT},
			{Name: "bytecodeIndex", Type: T_INT},
			{Name: "type", Type: T_FRAME_TYPE, ConstantPool: true},
		}},
		{Name: "jdk.types.Method", ID: T_METHOD, Fields: []def.Field{
			{Name: "type", Type: T_CLASS, ConstantPool: true},
			{Name: "name", Type: T_SYMBOL, ConstantPool: true},
			{Name: "descriptor", Type: T_SYMBOL, ConstantPool: true},
			{Name: "modifiers", Type: T_INT},
			{Name: "hidden", Type: T_BOOLEAN},
		}},
		{Name: "jdk.types.Package", ID: T_PACKAGE, Fields: []def.Field{
			{Name: "name", Type: T_SYMBOL, ConstantPool: true},
		}},
		{Name: "jdk.types.Symbol", ID: T_SYMBOL, Fields: []def.Field{
			{Name: "string", Type: T_STRING},
		}},
		{Name: "profiler.types.LogLevel", ID: T_LOG_LEVEL, Fields: []def.Field{
			{Name: "name", Type: T_STRING},
		}},

		// Event types
		{Name: "jdk.ExecutionSample", ID: T_EXECUTION_SAMPLE, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "sampledThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "state", Type: T_THREAD_STATE, ConstantPool: true},
			{Name: "spanId", Type: T_LONG},
			{Name: "spanName", Type: T_LONG},
			{Name: "contextId", Type: T_LONG},
		}},
		{Name: "profiler.WallClockSample", ID: T_WALL_CLOCK_SAMPLE, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "sampledThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "state", Type: T_THREAD_STATE, ConstantPool: true},
			{Name: "spanId", Type: T_LONG},
			{Name: "spanName", Type: T_LONG},
			{Name: "contextId", Type: T_LONG},
			{Name: "samples", Type: T_INT},
		}},
		{Name: "jdk.ObjectAllocationInNewTLAB", ID: T_ALLOC_IN_NEW_TLAB, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "objectClass", Type: T_CLASS, ConstantPool: true},
			{Name: "allocationSize", Type: T_LONG},
			{Name: "tlabSize", Type: T_LONG},
			{Name: "contextId", Type: T_LONG},
			{Name: "spanId", Type: T_LONG},
			{Name: "spanName", Type: T_LONG},
		}},
		{Name: "jdk.ObjectAllocationOutsideTLAB", ID: T_ALLOC_OUTSIDE_TLAB, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "objectClass", Type: T_CLASS, ConstantPool: true},
			{Name: "allocationSize", Type: T_LONG},
			{Name: "contextId", Type: T_LONG},
			{Name: "spanId", Type: T_LONG},
			{Name: "spanName", Type: T_LONG},
		}},
		{Name: "jdk.ObjectAllocationSample", ID: T_ALLOC_SAMPLE, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "objectClass", Type: T_CLASS, ConstantPool: true},
			{Name: "weight", Type: T_LONG},
		}},
		{Name: "jdk.JavaMonitorEnter", ID: T_MONITOR_ENTER, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "duration", Type: T_LONG},
			{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "monitorClass", Type: T_CLASS, ConstantPool: true},
			{Name: "previousOwner", Type: T_THREAD, ConstantPool: true},
			{Name: "address", Type: T_LONG},
			{Name: "contextId", Type: T_LONG},
			{Name: "spanId", Type: T_LONG},
			{Name: "spanName", Type: T_LONG},
		}},
		{Name: "jdk.ThreadPark", ID: T_THREAD_PARK, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "duration", Type: T_LONG},
			{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "parkedClass", Type: T_CLASS, ConstantPool: true},
			{Name: "timeout", Type: T_LONG},
			{Name: "until", Type: T_LONG},
			{Name: "address", Type: T_LONG},
		}},
		{Name: "profiler.LiveObject", ID: T_LIVE_OBJECT, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "objectClass", Type: T_CLASS, ConstantPool: true},
			{Name: "allocationSize", Type: T_LONG},
			{Name: "allocationTime", Type: T_LONG},
		}},
		{Name: "jdk.ActiveSetting", ID: T_ACTIVE_SETTING, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "duration", Type: T_LONG},
			{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "id", Type: T_LONG},
			{Name: "name", Type: T_STRING},
			{Name: "value", Type: T_STRING},
		}},
		{Name: "profiler.Malloc", ID: T_MALLOC, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "address", Type: T_LONG},
			{Name: "size", Type: T_LONG},
		}},
		{Name: "profiler.Free", ID: T_FREE, Fields: []def.Field{
			{Name: "startTime", Type: T_LONG},
			{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
			{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
			{Name: "address", Type: T_LONG},
		}},

		// Annotation types
		{Name: "jdk.jfr.Label", ID: T_LABEL, Fields: []def.Field{
			{Name: "value", Type: T_STRING},
		}},
		{Name: "jdk.jfr.Category", ID: T_CATEGORY, Fields: []def.Field{
			{Name: "value", Type: T_STRING, Array: true},
		}},
		{Name: "jdk.jfr.Timestamp", ID: T_TIMESTAMP, Fields: []def.Field{
			{Name: "value", Type: T_STRING},
		}},
		{Name: "jdk.jfr.Timespan", ID: T_TIMESPAN, Fields: []def.Field{
			{Name: "value", Type: T_STRING},
		}},
		{Name: "jdk.jfr.DataAmount", ID: T_DATA_AMOUNT, Fields: []def.Field{
			{Name: "value", Type: T_STRING},
		}},
		{Name: "jdk.jfr.MemoryAddress", ID: T_MEMORY_ADDRESS},
		{Name: "jdk.jfr.Unsigned", ID: T_UNSIGNED},
		{Name: "jdk.jfr.Percentage", ID: T_PERCENTAGE},
	}
}

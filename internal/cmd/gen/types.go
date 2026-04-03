package main

import (
	"fmt"

	"github.com/grafana/jfr-parser/parser/types"
)

const (
	T_METADATA types.TypeID = iota
	T_CPOOL
	T_BOOLEAN
	T_CHAR
	T_FLOAT
	T_DOUBLE
	T_BYTE
	T_SHORT
	T_INT
	T_LONG
	T_STRING
	T_CLASS
	T_THREAD
	T_CLASS_LOADER
	T_FRAME_TYPE
	T_THREAD_STATE
	T_STACK_TRACE
	T_STACK_FRAME
	T_METHOD
	T_PACKAGE
	T_SYMBOL
	T_LOG_LEVEL
	T_EVENT
	T_EXECUTION_SAMPLE
	T_ALLOC_IN_NEW_TLAB
	T_ALLOC_OUTSIDE_TLAB
	T_MONITOR_ENTER
	T_THREAD_PARK
	T_CPU_LOAD
	T_ACTIVE_RECORDING
	T_ACTIVE_SETTING
	T_OS_INFORMATION
	T_CPU_INFORMATION
	T_JVM_INFORMATION
	T_INITIAL_SYSTEM_PROPERTY
	T_NATIVE_LIBRARY
	T_LOG
	T_LIVE_OBJECT
	T_WALL_CLOCK_SAMPLE
	T_MALLOC
	T_FREE
	T_ANNOTATION
	T_LABEL
	T_CATEGORY
	T_TIMESTAMP
	T_TIMESPAN
	T_DATA_AMOUNT
	T_MEMORY_ADDRESS
	T_UNSIGNED
	T_PERCENTAGE
	T_ALLOC_SAMPLE
)

func TypeID2Sym(id types.TypeID) string {
	switch id {
	case T_METADATA:
		return "T_METADATA"
	case T_CPOOL:
		return "T_CPOOL"
	case T_BOOLEAN:
		return "T_BOOLEAN"
	case T_CHAR:
		return "T_CHAR"
	case T_FLOAT:
		return "T_FLOAT"
	case T_DOUBLE:
		return "T_DOUBLE"
	case T_BYTE:
		return "T_BYTE"
	case T_SHORT:
		return "T_SHORT"
	case T_INT:
		return "T_INT"
	case T_LONG:
		return "T_LONG"
	case T_STRING:
		return "T_STRING"
	case T_CLASS:
		return "T_CLASS"
	case T_THREAD:
		return "T_THREAD"
	case T_CLASS_LOADER:
		return "T_CLASS_LOADER"
	case T_FRAME_TYPE:
		return "T_FRAME_TYPE"
	case T_THREAD_STATE:
		return "T_THREAD_STATE"
	case T_STACK_TRACE:
		return "T_STACK_TRACE"
	case T_STACK_FRAME:
		return "T_STACK_FRAME"
	case T_METHOD:
		return "T_METHOD"
	case T_PACKAGE:
		return "T_PACKAGE"
	case T_SYMBOL:
		return "T_SYMBOL"
	case T_LOG_LEVEL:
		return "T_LOG_LEVEL"
	case T_EVENT:
		return "T_EVENT"
	case T_EXECUTION_SAMPLE:
		return "T_EXECUTION_SAMPLE"
	case T_ALLOC_IN_NEW_TLAB:
		return "T_ALLOC_IN_NEW_TLAB"
	case T_ALLOC_OUTSIDE_TLAB:
		return "T_ALLOC_OUTSIDE_TLAB"
	case T_ALLOC_SAMPLE:
		return "T_ALLOC_SAMPLE"
	case T_MONITOR_ENTER:
		return "T_MONITOR_ENTER"
	case T_THREAD_PARK:
		return "T_THREAD_PARK"
	case T_CPU_LOAD:
		return "T_CPU_LOAD"
	case T_ACTIVE_RECORDING:
		return "T_ACTIVE_RECORDING"
	case T_ACTIVE_SETTING:
		return "T_ACTIVE_SETTING"
	case T_OS_INFORMATION:
		return "T_OS_INFORMATION"
	case T_CPU_INFORMATION:
		return "T_CPU_INFORMATION"
	case T_JVM_INFORMATION:
		return "T_JVM_INFORMATION"
	case T_INITIAL_SYSTEM_PROPERTY:
		return "T_INITIAL_SYSTEM_PROPERTY"
	case T_NATIVE_LIBRARY:
		return "T_NATIVE_LIBRARY"
	case T_LOG:
		return "T_LOG"
	case T_LIVE_OBJECT:
		return "T_LIVE_OBJECT"
	case T_ANNOTATION:
		return "T_ANNOTATION"
	case T_LABEL:
		return "T_LABEL"
	case T_CATEGORY:
		return "T_CATEGORY"
	case T_TIMESTAMP:
		return "T_TIMESTAMP"
	case T_TIMESPAN:
		return "T_TIMESPAN"
	case T_DATA_AMOUNT:
		return "T_DATA_AMOUNT"
	case T_MEMORY_ADDRESS:
		return "T_MEMORY_ADDRESS"
	case T_UNSIGNED:
		return "T_UNSIGNED"
	case T_PERCENTAGE:
		return "T_PERCENTAGE"
	default:
		return fmt.Sprintf("unknown type %d", id)
	}
}

var Type_boolean = types.MetadataClass{
	Name:   "boolean",
	ID:     T_BOOLEAN,
	Fields: []types.Field{},
}
var Type_char = types.MetadataClass{
	Name:   "char",
	ID:     T_CHAR,
	Fields: []types.Field{},
}
var Type_float = types.MetadataClass{
	Name:   "float",
	ID:     T_FLOAT,
	Fields: []types.Field{},
}
var Type_double = types.MetadataClass{
	Name:   "double",
	ID:     T_DOUBLE,
	Fields: []types.Field{},
}
var Type_byte = types.MetadataClass{
	Name:   "byte",
	ID:     T_BYTE,
	Fields: []types.Field{},
}
var Type_short = types.MetadataClass{
	Name:   "short",
	ID:     T_SHORT,
	Fields: []types.Field{},
}
var Type_int = types.MetadataClass{
	Name:   "int",
	ID:     T_INT,
	Fields: []types.Field{},
}
var Type_long = types.MetadataClass{
	Name:   "long",
	ID:     T_LONG,
	Fields: []types.Field{},
}
var Type_java_lang_String = types.MetadataClass{
	Name:   "java.lang.String",
	ID:     T_STRING,
	Fields: []types.Field{},
}
var Type_java_lang_Class = types.MetadataClass{
	Name: "java.lang.Class",
	ID:   T_CLASS,
	Fields: []types.Field{
		{Name: "classLoader", Type: T_CLASS_LOADER, ConstantPool: true},
		{Name: "name", Type: T_SYMBOL, ConstantPool: true},
		{Name: "package", Type: T_PACKAGE, ConstantPool: true},
		{Name: "modifiers", Type: T_INT, ConstantPool: false},
	},
}
var Type_java_lang_Thread = types.MetadataClass{
	Name: "java.lang.Thread",
	ID:   T_THREAD,
	Fields: []types.Field{
		{Name: "osName", Type: T_STRING, ConstantPool: false},
		{Name: "osThreadId", Type: T_LONG, ConstantPool: false},
		{Name: "javaName", Type: T_STRING, ConstantPool: false},
		{Name: "javaThreadId", Type: T_LONG, ConstantPool: false},
	},
}
var Type_jdk_types_ClassLoader = types.MetadataClass{
	Name: "jdk.types.ClassLoader",
	ID:   T_CLASS_LOADER,
	Fields: []types.Field{
		{Name: "type", Type: T_CLASS, ConstantPool: true},
		{Name: "name", Type: T_SYMBOL, ConstantPool: true},
	},
}
var Type_jdk_types_FrameType = types.MetadataClass{
	Name: "jdk.types.FrameType",
	ID:   T_FRAME_TYPE,
	Fields: []types.Field{
		{Name: "description", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_types_ThreadState = types.MetadataClass{
	Name: "jdk.types.ThreadState",
	ID:   T_THREAD_STATE,
	Fields: []types.Field{
		{Name: "name", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_types_StackTrace = types.MetadataClass{
	Name: "jdk.types.StackTrace",
	ID:   T_STACK_TRACE,
	Fields: []types.Field{
		{Name: "truncated", Type: T_BOOLEAN, ConstantPool: false},
		{Name: "frames", Type: T_STACK_FRAME, ConstantPool: false, Array: true},
	},
}
var Type_jdk_types_StackFrame = types.MetadataClass{
	Name: "jdk.types.StackFrame",
	ID:   T_STACK_FRAME,
	Fields: []types.Field{
		{Name: "method", Type: T_METHOD, ConstantPool: true},
		{Name: "lineNumber", Type: T_INT, ConstantPool: false},
		{Name: "bytecodeIndex", Type: T_INT, ConstantPool: false},
		{Name: "type", Type: T_FRAME_TYPE, ConstantPool: true},
	},
}
var Type_jdk_types_Method = types.MetadataClass{
	Name: "jdk.types.Method",
	ID:   T_METHOD,
	Fields: []types.Field{
		{Name: "type", Type: T_CLASS, ConstantPool: true},
		{Name: "name", Type: T_SYMBOL, ConstantPool: true},
		{Name: "descriptor", Type: T_SYMBOL, ConstantPool: true},
		{Name: "modifiers", Type: T_INT, ConstantPool: false},
		{Name: "hidden", Type: T_BOOLEAN, ConstantPool: false},
	},
}
var Type_jdk_types_Package = types.MetadataClass{
	Name: "jdk.types.Package",
	ID:   T_PACKAGE,
	Fields: []types.Field{
		{Name: "name", Type: T_SYMBOL, ConstantPool: true},
	},
}
var Type_jdk_types_Symbol = types.MetadataClass{
	Name: "jdk.types.Symbol",
	ID:   T_SYMBOL,
	Fields: []types.Field{
		{Name: "string", Type: T_STRING, ConstantPool: false},
	},
}
var Type_profiler_types_LogLevel = types.MetadataClass{
	Name: "profiler.types.LogLevel",
	ID:   T_LOG_LEVEL,
	Fields: []types.Field{
		{Name: "name", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_ExecutionSample = types.MetadataClass{
	Name: "jdk.ExecutionSample",
	ID:   T_EXECUTION_SAMPLE,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "sampledThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "state", Type: T_THREAD_STATE, ConstantPool: true},
		{Name: "spanId", Type: T_LONG, ConstantPool: false},
		{Name: "spanName", Type: T_LONG, ConstantPool: false},
		{Name: "contextId", Type: T_LONG, ConstantPool: false},
	},
}

var Type_profiler_WallClockSample = types.MetadataClass{
	Name: "profiler.WallClockSample",
	ID:   T_WALL_CLOCK_SAMPLE,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "sampledThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "state", Type: T_THREAD_STATE, ConstantPool: true},
		{Name: "spanId", Type: T_LONG, ConstantPool: false},
		{Name: "spanName", Type: T_LONG, ConstantPool: false},
		{Name: "contextId", Type: T_LONG, ConstantPool: false},
		{Name: "samples", Type: T_INT, ConstantPool: false},
	},
}

var Type_jdk_ObjectAllocationInNewTLAB = types.MetadataClass{
	Name: "jdk.ObjectAllocationInNewTLAB",
	ID:   T_ALLOC_IN_NEW_TLAB,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "objectClass", Type: T_CLASS, ConstantPool: true},
		{Name: "allocationSize", Type: T_LONG, ConstantPool: false},
		{Name: "tlabSize", Type: T_LONG, ConstantPool: false},
		{Name: "contextId", Type: T_LONG, ConstantPool: false},
		{Name: "spanId", Type: T_LONG, ConstantPool: false},
		{Name: "spanName", Type: T_LONG, ConstantPool: false},
	},
}
var Type_jdk_ObjectAllocationOutsideTLAB = types.MetadataClass{
	Name: "jdk.ObjectAllocationOutsideTLAB",
	ID:   T_ALLOC_OUTSIDE_TLAB,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "objectClass", Type: T_CLASS, ConstantPool: true},
		{Name: "allocationSize", Type: T_LONG, ConstantPool: false},
		{Name: "contextId", Type: T_LONG, ConstantPool: false},
		{Name: "spanId", Type: T_LONG, ConstantPool: false},
		{Name: "spanName", Type: T_LONG, ConstantPool: false},
	},
}
var Type_jdk_ObjectAllocationSample = types.MetadataClass{
	Name: "jdk.ObjectAllocationSample",
	ID:   T_ALLOC_SAMPLE,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "objectClass", Type: T_CLASS, ConstantPool: true},
		{Name: "weight", Type: T_LONG, ConstantPool: false},
	},
}
var Type_jdk_JavaMonitorEnter = types.MetadataClass{
	Name: "jdk.JavaMonitorEnter",
	ID:   T_MONITOR_ENTER,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "duration", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "monitorClass", Type: T_CLASS, ConstantPool: true},
		{Name: "previousOwner", Type: T_THREAD, ConstantPool: true},
		{Name: "address", Type: T_LONG, ConstantPool: false},
		{Name: "contextId", Type: T_LONG, ConstantPool: false},
		{Name: "spanId", Type: T_LONG, ConstantPool: false},
		{Name: "spanName", Type: T_LONG, ConstantPool: false},
	},
}
var Type_jdk_ThreadPark = types.MetadataClass{
	Name: "jdk.ThreadPark",
	ID:   T_THREAD_PARK,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "duration", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "parkedClass", Type: T_CLASS, ConstantPool: true},
		{Name: "timeout", Type: T_LONG, ConstantPool: false},
		{Name: "until", Type: T_LONG, ConstantPool: false},
		{Name: "address", Type: T_LONG, ConstantPool: false},
	},
}
var Type_jdk_CPULoad = types.MetadataClass{
	Name: "jdk.CPULoad",
	ID:   T_CPU_LOAD,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "jvmUser", Type: T_FLOAT, ConstantPool: false},
		{Name: "jvmSystem", Type: T_FLOAT, ConstantPool: false},
		{Name: "machineTotal", Type: T_FLOAT, ConstantPool: false},
	},
}
var Type_jdk_ActiveRecording = types.MetadataClass{
	Name: "jdk.ActiveRecording",
	ID:   T_ACTIVE_RECORDING,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "duration", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "id", Type: T_LONG, ConstantPool: false},
		{Name: "name", Type: T_STRING, ConstantPool: false},
		{Name: "destination", Type: T_STRING, ConstantPool: false},
		{Name: "maxAge", Type: T_LONG, ConstantPool: false},
		{Name: "maxSize", Type: T_LONG, ConstantPool: false},
		{Name: "recordingStart", Type: T_LONG, ConstantPool: false},
		{Name: "recordingDuration", Type: T_LONG, ConstantPool: false},
	},
}
var Type_jdk_ActiveSetting = types.MetadataClass{
	Name: "jdk.ActiveSetting",
	ID:   T_ACTIVE_SETTING,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "duration", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "id", Type: T_LONG, ConstantPool: false},
		{Name: "name", Type: T_STRING, ConstantPool: false},
		{Name: "value", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_OSInformation = types.MetadataClass{
	Name: "jdk.OSInformation",
	ID:   T_OS_INFORMATION,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "osVersion", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_CPUInformation = types.MetadataClass{
	Name: "jdk.CPUInformation",
	ID:   T_CPU_INFORMATION,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "cpu", Type: T_STRING, ConstantPool: false},
		{Name: "description", Type: T_STRING, ConstantPool: false},
		{Name: "sockets", Type: T_INT, ConstantPool: false},
		{Name: "cores", Type: T_INT, ConstantPool: false},
		{Name: "hwThreads", Type: T_INT, ConstantPool: false},
	},
}
var Type_jdk_JVMInformation = types.MetadataClass{
	Name: "jdk.JVMInformation",
	ID:   T_JVM_INFORMATION,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "jvmName", Type: T_STRING, ConstantPool: false},
		{Name: "jvmVersion", Type: T_STRING, ConstantPool: false},
		{Name: "jvmArguments", Type: T_STRING, ConstantPool: false},
		{Name: "jvmFlags", Type: T_STRING, ConstantPool: false},
		{Name: "javaArguments", Type: T_STRING, ConstantPool: false},
		{Name: "jvmStartTime", Type: T_LONG, ConstantPool: false},
		{Name: "pid", Type: T_LONG, ConstantPool: false},
	},
}
var Type_jdk_InitialSystemProperty = types.MetadataClass{
	Name: "jdk.InitialSystemProperty",
	ID:   T_INITIAL_SYSTEM_PROPERTY,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "key", Type: T_STRING, ConstantPool: false},
		{Name: "value", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_NativeLibrary = types.MetadataClass{
	Name: "jdk.NativeLibrary",
	ID:   T_NATIVE_LIBRARY,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "name", Type: T_STRING, ConstantPool: false},
		{Name: "baseAddress", Type: T_LONG, ConstantPool: false},
		{Name: "topAddress", Type: T_LONG, ConstantPool: false},
	},
}
var Type_profiler_Log = types.MetadataClass{
	Name: "profiler.Log",
	ID:   T_LOG,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "level", Type: T_LOG_LEVEL, ConstantPool: true},
		{Name: "message", Type: T_STRING, ConstantPool: false},
	},
}
var Type_profiler_LiveObject = types.MetadataClass{
	Name: "profiler.LiveObject",
	ID:   T_LIVE_OBJECT,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "objectClass", Type: T_CLASS, ConstantPool: true},
		{Name: "allocationSize", Type: T_LONG, ConstantPool: false},
		{Name: "allocationTime", Type: T_LONG, ConstantPool: false},
	},
}
var Type_jdk_jfr_Label = types.MetadataClass{
	Name: "jdk.jfr.Label",
	ID:   T_LABEL,
	Fields: []types.Field{
		{Name: "value", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_jfr_Category = types.MetadataClass{
	Name: "jdk.jfr.Category",
	ID:   T_CATEGORY,
	Fields: []types.Field{
		{Name: "value", Type: T_STRING, ConstantPool: false, Array: true},
	},
}
var Type_jdk_jfr_Timestamp = types.MetadataClass{
	Name: "jdk.jfr.Timestamp",
	ID:   T_TIMESTAMP,
	Fields: []types.Field{
		{Name: "value", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_jfr_Timespan = types.MetadataClass{
	Name: "jdk.jfr.Timespan",
	ID:   T_TIMESPAN,
	Fields: []types.Field{
		{Name: "value", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_jfr_DataAmount = types.MetadataClass{
	Name: "jdk.jfr.DataAmount",
	ID:   T_DATA_AMOUNT,
	Fields: []types.Field{
		{Name: "value", Type: T_STRING, ConstantPool: false},
	},
}
var Type_jdk_jfr_MemoryAddress = types.MetadataClass{
	Name:   "jdk.jfr.MemoryAddress",
	ID:     T_MEMORY_ADDRESS,
	Fields: []types.Field{},
}
var Type_jdk_jfr_Unsigned = types.MetadataClass{
	Name:   "jdk.jfr.Unsigned",
	ID:     T_UNSIGNED,
	Fields: []types.Field{},
}
var Type_jdk_jfr_Percentage = types.MetadataClass{
	Name:   "jdk.jfr.Percentage",
	ID:     T_PERCENTAGE,
	Fields: []types.Field{},
}

var Type_profiler_Malloc = types.MetadataClass{
	Name: "profiler.Malloc",
	ID:   T_MALLOC,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "address", Type: T_LONG, ConstantPool: false},
		{Name: "size", Type: T_LONG, ConstantPool: false},
	},
}

var Type_profiler_Free = types.MetadataClass{
	Name: "profiler.Free",
	ID:   T_FREE,
	Fields: []types.Field{
		{Name: "startTime", Type: T_LONG, ConstantPool: false},
		{Name: "eventThread", Type: T_THREAD, ConstantPool: true},
		{Name: "stackTrace", Type: T_STACK_TRACE, ConstantPool: true},
		{Name: "address", Type: T_LONG, ConstantPool: false},
	},
}

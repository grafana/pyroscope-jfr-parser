package jfrwriter

// SerializeSymbol serializes a Symbol constant pool entry: string field.
func SerializeSymbol(s string) []byte {
	var w Writer
	w.String(s)
	return w.Bytes()
}

// SerializeFrameType serializes a FrameType constant pool entry: description:string.
func SerializeFrameType(description string) []byte {
	var w Writer
	w.String(description)
	return w.Bytes()
}

// SerializeThreadState serializes a ThreadState constant pool entry: name:string.
func SerializeThreadState(name string) []byte {
	var w Writer
	w.String(name)
	return w.Bytes()
}

// SerializeThread serializes a Thread constant pool entry.
// Fields: osName:string, osThreadId:long, javaName:string, javaThreadId:long.
func SerializeThread(osName string, osThreadId uint64, javaName string, javaThreadId uint64) []byte {
	var w Writer
	w.String(osName)
	w.VarLong(osThreadId)
	w.String(javaName)
	w.VarLong(javaThreadId)
	return w.Bytes()
}

// SerializeClassLoader serializes a ClassLoader constant pool entry.
// Fields: type:ClassRef(CP), name:SymbolRef(CP).
func SerializeClassLoader(typeRef, nameRef uint64) []byte {
	var w Writer
	w.VarLong(typeRef)
	w.VarLong(nameRef)
	return w.Bytes()
}

// SerializeClass serializes a Class constant pool entry.
// Fields: classLoader:ClassLoaderRef(CP), name:SymbolRef(CP), package:PackageRef(CP), modifiers:int.
func SerializeClass(classLoaderRef, nameRef, packageRef uint64, modifiers uint32) []byte {
	var w Writer
	w.VarLong(classLoaderRef)
	w.VarLong(nameRef)
	w.VarLong(packageRef)
	w.VarInt(modifiers)
	return w.Bytes()
}

// SerializeMethod serializes a Method constant pool entry.
// Fields: type:ClassRef(CP), name:SymbolRef(CP), descriptor:SymbolRef(CP), modifiers:int, hidden:boolean.
func SerializeMethod(typeRef, nameRef, descriptorRef uint64, modifiers uint32, hidden bool) []byte {
	var w Writer
	w.VarLong(typeRef)
	w.VarLong(nameRef)
	w.VarLong(descriptorRef)
	w.VarInt(modifiers)
	w.Bool(hidden)
	return w.Bytes()
}

// SerializePackage serializes a Package constant pool entry.
// Fields: name:SymbolRef(CP).
func SerializePackage(nameRef uint64) []byte {
	var w Writer
	w.VarLong(nameRef)
	return w.Bytes()
}

// SerializeStackFrame serializes a single StackFrame (inline struct, not a cpool entry).
// Fields: method:MethodRef(CP), lineNumber:int, bytecodeIndex:int, type:FrameTypeRef(CP).
func SerializeStackFrame(methodRef uint64, lineNumber, bytecodeIndex uint32, frameTypeRef uint64) []byte {
	var w Writer
	w.VarLong(methodRef)
	w.VarInt(lineNumber)
	w.VarInt(bytecodeIndex)
	w.VarLong(frameTypeRef)
	return w.Bytes()
}

// SerializeStackTrace serializes a StackTrace constant pool entry.
// Fields: truncated:boolean, frames:StackFrame[].
// Each frame should be pre-serialized with SerializeStackFrame.
func SerializeStackTrace(truncated bool, frames [][]byte) []byte {
	var w Writer
	w.Bool(truncated)
	w.VarInt(uint32(len(frames)))
	for _, f := range frames {
		w.Raw(f)
	}
	return w.Bytes()
}

// SerializeLogLevel serializes a LogLevel constant pool entry: name:string.
func SerializeLogLevel(name string) []byte {
	var w Writer
	w.String(name)
	return w.Bytes()
}

// SerializeExecutionSample serializes an ExecutionSample event body.
// Fields: startTime:long, sampledThread:ThreadRef(CP), stackTrace:StackTraceRef(CP),
// state:ThreadStateRef(CP), spanId:long, spanName:long, contextId:long.
func SerializeExecutionSample(startTime, threadRef, stackTraceRef, stateRef uint64) []byte {
	var w Writer
	w.VarLong(startTime)
	w.VarLong(threadRef)
	w.VarLong(stackTraceRef)
	w.VarLong(stateRef)
	w.VarLong(0) // spanId
	w.VarLong(0) // spanName
	w.VarLong(0) // contextId
	return w.Bytes()
}

// SerializeObjectAllocationSample serializes an ObjectAllocationSample event body.
// Fields: startTime:long, eventThread:ThreadRef(CP), stackTrace:StackTraceRef(CP),
// objectClass:ClassRef(CP), weight:long.
func SerializeObjectAllocationSample(startTime, threadRef, stackTraceRef, classRef, weight uint64) []byte {
	var w Writer
	w.VarLong(startTime)
	w.VarLong(threadRef)
	w.VarLong(stackTraceRef)
	w.VarLong(classRef)
	w.VarLong(weight)
	return w.Bytes()
}

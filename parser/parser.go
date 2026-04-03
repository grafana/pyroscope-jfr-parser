package parser

import (
	"fmt"
	"io"
	"unsafe"

	types2 "github.com/grafana/jfr-parser/parser/types"
	"github.com/grafana/jfr-parser/parser/types/def"
)

const chunkHeaderSize = 68
const bufferSize = 1024 * 1024
const chunkMagic = 0x464c5200

type ChunkHeader struct {
	Magic              uint32
	Version            uint32
	Size               int
	OffsetConstantPool int
	OffsetMeta         int
	StartNanos         uint64
	DurationNanos      uint64
	StartTicks         uint64
	TicksPerSecond     uint64
	Features           uint32
}

func (c *ChunkHeader) String() string {
	return fmt.Sprintf("ChunkHeader{Magic: %x, Version: %x, Size: %d, OffsetConstantPool: %d, OffsetMeta: %d, StartNanos: %d, DurationNanos: %d, StartTicks: %d, TicksPerSecond: %d, Features: %d}", c.Magic, c.Version, c.Size, c.OffsetConstantPool, c.OffsetMeta, c.StartNanos, c.DurationNanos, c.StartTicks, c.TicksPerSecond, c.Features)
}

type SymbolProcessor func(ref *types2.SymbolList)

type Options struct {
	ChunkSizeLimit  int
	SymbolProcessor SymbolProcessor
}

type Parser struct {
	FrameTypes   types2.FrameTypeList
	ThreadStates types2.ThreadStateList
	Threads      types2.ThreadList
	Classes      types2.ClassList
	Methods      types2.MethodList
	Packages     types2.PackageList
	Symbols      types2.SymbolList
	LogLevels    types2.LogLevelList
	Stacktrace   types2.StackTraceList
	Strings      types2.StringList

	ExecutionSample             types2.ExecutionSample
	WallClockSample             types2.WallClockSample
	Malloc                      types2.Malloc
	Free                        types2.Free
	ObjectAllocationInNewTLAB   types2.ObjectAllocationInNewTLAB
	ObjectAllocationOutsideTLAB types2.ObjectAllocationOutsideTLAB
	ObjectAllocationSample      types2.ObjectAllocationSample
	JavaMonitorEnter            types2.JavaMonitorEnter
	ThreadPark                  types2.ThreadPark
	LiveObject                  types2.LiveObject
	ActiveSetting               types2.ActiveSetting

	header   ChunkHeader
	options  Options
	buf      []byte
	pos      int
	metaSize uint32
	chunkEnd int

	TypeMap def.TypeMap

	bindFrameType   *types2.BindFrameType
	bindThreadState *types2.BindThreadState
	bindThread      *types2.BindThread
	bindClass       *types2.BindClass
	bindMethod      *types2.BindMethod
	bindPackage     *types2.BindPackage
	bindSymbol      *types2.BindSymbol
	bindLogLevel    *types2.BindLogLevel
	bindStackFrame  *types2.BindStackFrame
	bindStackTrace  *types2.BindStackTrace
	bindString      *types2.BindString

	bindExecutionSample *types2.BindExecutionSample

	bindAllocInNewTLAB   *types2.BindObjectAllocationInNewTLAB
	bindAllocOutsideTLAB *types2.BindObjectAllocationOutsideTLAB
	bindAllocSample      *types2.BindObjectAllocationSample
	bindMonitorEnter     *types2.BindJavaMonitorEnter
	bindThreadPark       *types2.BindThreadPark
	bindLiveObject       *types2.BindLiveObject
	bindActiveSetting    *types2.BindActiveSetting
	bindWallClockSample  *types2.BindWallClockSample
	bindMalloc           *types2.BindMalloc
	bindFree             *types2.BindFree
}

func NewParser(buf []byte, options Options) *Parser {
	p := &Parser{
		options: options,
		buf:     buf,
	}
	return p
}

func initEventType[B any](
	cls *def.Class, typeID *def.TypeID, bind **B,
	newBind func(*def.Class, *def.TypeMap) *B, typeMap *def.TypeMap,
) {
	if cls != nil {
		*typeID = cls.ID
		*bind = newBind(cls, typeMap)
	} else {
		*typeID = def.UnsetTypeID
		*bind = nil
	}
}

func (p *Parser) ParseEvent() (def.TypeID, error) {
	for {
		if p.pos == p.chunkEnd {
			if p.pos == len(p.buf) {
				return 0, io.EOF
			}
			if err := p.readChunk(p.pos); err != nil {
				return 0, err
			}
		}
		pp := p.pos
		size, err := p.varLong()
		if err != nil {
			return 0, err
		}
		if size == 0 {
			return 0, def.ErrIntOverflow
		}
		if uint64(p.chunkEnd-pp) < size {
			return 0, fmt.Errorf("invalid event size %d at position %d", size, pp)
		}
		typ, err := p.varLong()
		if err != nil {
			return 0, err
		}
		_ = size

		ttyp := def.TypeID(typ)
		switch ttyp {
		case def.UnsetTypeID:
			return 0, fmt.Errorf("invalid event type %d at position %d", typ, pp)
		case p.TypeMap.T_EXECUTION_SAMPLE:
			_, err = p.ExecutionSample.Parse(p.buf[p.pos:], p.bindExecutionSample, &p.TypeMap)
		case p.TypeMap.T_WALL_CLOCK_SAMPLE:
			_, err = p.WallClockSample.Parse(p.buf[p.pos:], p.bindWallClockSample, &p.TypeMap)
		case p.TypeMap.T_MALLOC:
			_, err = p.Malloc.Parse(p.buf[p.pos:], p.bindMalloc, &p.TypeMap)
		case p.TypeMap.T_FREE:
			_, err = p.Free.Parse(p.buf[p.pos:], p.bindFree, &p.TypeMap)
		case p.TypeMap.T_ALLOC_IN_NEW_TLAB:
			_, err = p.ObjectAllocationInNewTLAB.Parse(p.buf[p.pos:], p.bindAllocInNewTLAB, &p.TypeMap)
		case p.TypeMap.T_ALLOC_OUTSIDE_TLAB:
			_, err = p.ObjectAllocationOutsideTLAB.Parse(p.buf[p.pos:], p.bindAllocOutsideTLAB, &p.TypeMap)
		case p.TypeMap.T_ALLOC_SAMPLE:
			_, err = p.ObjectAllocationSample.Parse(p.buf[p.pos:], p.bindAllocSample, &p.TypeMap)
		case p.TypeMap.T_LIVE_OBJECT:
			_, err = p.LiveObject.Parse(p.buf[p.pos:], p.bindLiveObject, &p.TypeMap)
		case p.TypeMap.T_MONITOR_ENTER:
			_, err = p.JavaMonitorEnter.Parse(p.buf[p.pos:], p.bindMonitorEnter, &p.TypeMap)
		case p.TypeMap.T_THREAD_PARK:
			_, err = p.ThreadPark.Parse(p.buf[p.pos:], p.bindThreadPark, &p.TypeMap)
		case p.TypeMap.T_ACTIVE_SETTING:
			_, err = p.ActiveSetting.Parse(p.buf[p.pos:], p.bindActiveSetting, &p.TypeMap)
		default:
			p.pos = pp + int(size)
			continue
		}
		if err != nil {
			return 0, err
		}
		p.pos = pp + int(size)
		return ttyp, nil
	}
}

func (p *Parser) ChunkHeader() ChunkHeader {
	return p.header
}

func (p *Parser) GetStacktrace(stID types2.StackTraceRef) *types2.StackTrace {
	idx, ok := p.Stacktrace.IDMap[stID]
	if !ok {
		return nil
	}
	return &p.Stacktrace.StackTrace[idx]
}

func (p *Parser) GetThreadState(ref types2.ThreadStateRef) *types2.ThreadState {
	idx, ok := p.ThreadStates.IDMap[ref]
	if !ok {
		return nil
	}
	return &p.ThreadStates.ThreadState[idx]
}

func (p *Parser) GetMethod(mID types2.MethodRef) *types2.Method {
	idx, ok := p.Methods.IDMap[mID]
	if !ok || int(idx) >= len(p.Methods.Method) {
		return nil
	}
	return &p.Methods.Method[idx]
}

func (p *Parser) GetClass(cID types2.ClassRef) *types2.Class {
	idx, ok := p.Classes.IDMap[cID]
	if !ok {
		return nil
	}
	return &p.Classes.Class[idx]
}

func (p *Parser) GetSymbol(sID types2.SymbolRef) *types2.Symbol {
	idx, ok := p.Symbols.IDMap[sID]
	if !ok {
		return nil
	}
	return &p.Symbols.Symbol[idx]
}

func (p *Parser) GetSymbolString(sID types2.SymbolRef) string {
	idx, ok := p.Symbols.IDMap[sID]
	if !ok {
		return ""
	}
	return p.Symbols.Symbol[idx].String
}

func (p *Parser) readChunk(pos int) error {
	if err := p.readChunkHeader(pos); err != nil {
		return fmt.Errorf("error reading chunk header: %w", err)
	}

	if err := p.readMeta(pos + p.header.OffsetMeta); err != nil {
		return fmt.Errorf("error reading metadata: %w", err)
	}
	if err := p.readConstantPool(pos + p.header.OffsetConstantPool); err != nil {
		return fmt.Errorf("error reading CP: %w @ %d", err, pos+p.header.OffsetConstantPool)
	}
	pp := p.options.SymbolProcessor
	if pp != nil {
		pp(&p.Symbols)
	}
	p.pos = pos + chunkHeaderSize
	return nil
}

func (p *Parser) seek(pos int) error {
	if pos >= 0 && pos < len(p.buf) {
		p.pos = pos
		return nil
	}
	return io.ErrUnexpectedEOF
}

func (p *Parser) byte() (byte, error) {
	if p.pos >= len(p.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	b := p.buf[p.pos]
	p.pos++
	return b, nil
}
func (p *Parser) varInt() (uint32, error) {
	v := uint32(0)
	for shift := uint(0); ; shift += 7 {
		if shift >= 32 {
			return 0, def.ErrIntOverflow
		}
		if p.pos >= len(p.buf) {
			return 0, io.ErrUnexpectedEOF
		}
		b := p.buf[p.pos]
		p.pos++
		v |= uint32(b&0x7F) << shift
		if b < 0x80 {
			break
		}
	}
	return v, nil
}

func (p *Parser) varLong() (uint64, error) {
	v64_ := uint64(0)
	for shift := uint(0); shift <= 56; shift += 7 {
		if p.pos >= len(p.buf) {
			return 0, io.ErrUnexpectedEOF
		}
		b_ := p.buf[p.pos]
		p.pos++
		if shift == 56 {
			v64_ |= uint64(b_&0xFF) << shift
			break
		} else {
			v64_ |= uint64(b_&0x7F) << shift
			if b_ < 0x80 {
				break
			}
		}
	}
	return v64_, nil
}

func (p *Parser) string() (string, error) {
	if p.pos >= len(p.buf) {
		return "", io.ErrUnexpectedEOF
	}
	b := p.buf[p.pos]
	p.pos++
	switch b { //todo implement 2
	case 0:
		return "", nil //todo this should be nil
	case 1:
		return "", nil
	case 3:
		bs, err := p.bytes()
		if err != nil {
			return "", err
		}
		str := *(*string)(unsafe.Pointer(&bs))
		return str, nil
	case 4:
		return p.charArrayString()
	default:
		return "", fmt.Errorf("unknown string type %d", b)
	}

}

func (p *Parser) charArrayString() (string, error) {
	l, err := p.varInt()
	if err != nil {
		return "", err
	}
	if l < 0 {
		return "", def.ErrIntOverflow
	}
	buf := make([]rune, int(l))
	for i := 0; i < int(l); i++ {
		c, err := p.varInt()
		if err != nil {
			return "", err
		}
		buf[i] = rune(c)
	}

	res := string(buf)
	return res, nil
}

func (p *Parser) bytes() ([]byte, error) {
	l, err := p.varInt()
	if err != nil {
		return nil, err
	}
	if l < 0 {
		return nil, def.ErrIntOverflow
	}
	if p.pos+int(l) > len(p.buf) {
		return nil, io.ErrUnexpectedEOF
	}
	bs := p.buf[p.pos : p.pos+int(l)]
	p.pos += int(l)
	return bs, nil
}

func (p *Parser) checkTypes() error {

	tint := p.TypeMap.NameMap["int"]
	tlong := p.TypeMap.NameMap["long"]
	tshort := p.TypeMap.NameMap["short"]
	tfloat := p.TypeMap.NameMap["float"]
	tboolean := p.TypeMap.NameMap["boolean"]
	tstring := p.TypeMap.NameMap["java.lang.String"]

	if tint == nil {
		return fmt.Errorf("missing \"int\"")
	}
	if tlong == nil {
		return fmt.Errorf("missing \"long\"")
	}
	if tshort == nil {
		return fmt.Errorf("missing \"short\"")
	}
	if tfloat == nil {
		return fmt.Errorf("missing \"float\"")
	}
	if tboolean == nil {
		return fmt.Errorf("missing \"boolean\"")
	}
	if tstring == nil {
		return fmt.Errorf("missing \"java.lang.String\"")
	}
	p.TypeMap.T_INT = tint.ID
	p.TypeMap.T_LONG = tlong.ID
	p.TypeMap.T_SHORT = tshort.ID
	p.TypeMap.T_FLOAT = tfloat.ID
	p.TypeMap.T_BOOLEAN = tboolean.ID
	p.TypeMap.T_STRING = tstring.ID

	typeCPFrameType := p.TypeMap.NameMap["jdk.types.FrameType"]
	typeCPThreadState := p.TypeMap.NameMap["jdk.types.ThreadState"]
	typeCPThread := p.TypeMap.NameMap["java.lang.Thread"]
	typeCPClass := p.TypeMap.NameMap["java.lang.Class"]
	typeCPMethod := p.TypeMap.NameMap["jdk.types.Method"]
	typeCPPackage := p.TypeMap.NameMap["jdk.types.Package"]
	typeCPSymbol := p.TypeMap.NameMap["jdk.types.Symbol"]
	typeCPLogLevel := p.TypeMap.NameMap["profiler.types.LogLevel"]
	typeCPStackTrace := p.TypeMap.NameMap["jdk.types.StackTrace"]
	typeCPClassLoader := p.TypeMap.NameMap["jdk.types.ClassLoader"]

	if typeCPFrameType == nil {
		return fmt.Errorf("missing \"jdk.types.FrameType\"")
	}
	if typeCPThreadState == nil {
		return fmt.Errorf("missing \"jdk.types.ThreadState\"")
	}
	if typeCPThread == nil {
		return fmt.Errorf("missing \"java.lang.Thread\"")
	}
	if typeCPClass == nil {
		return fmt.Errorf("missing \"java.lang.Class\"")
	}
	if typeCPMethod == nil {
		return fmt.Errorf("missing \"jdk.types.Method\"")
	}
	if typeCPPackage == nil {
		return fmt.Errorf("missing \"jdk.types.Package\"")
	}
	if typeCPSymbol == nil {
		return fmt.Errorf("missing \"jdk.types.Symbol\"")
	}
	if typeCPStackTrace == nil {
		return fmt.Errorf("missing \"jdk.types.StackTrace\"")
	}
	if typeCPClassLoader == nil {
		return fmt.Errorf("missing \"jdk.types.ClassLoader\"")
	}
	p.TypeMap.T_FRAME_TYPE = typeCPFrameType.ID
	p.TypeMap.T_THREAD_STATE = typeCPThreadState.ID
	p.TypeMap.T_THREAD = typeCPThread.ID
	p.TypeMap.T_CLASS = typeCPClass.ID
	p.TypeMap.T_METHOD = typeCPMethod.ID
	p.TypeMap.T_PACKAGE = typeCPPackage.ID
	p.TypeMap.T_SYMBOL = typeCPSymbol.ID
	if typeCPLogLevel != nil {
		p.TypeMap.T_LOG_LEVEL = typeCPLogLevel.ID
	} else {
		p.TypeMap.T_LOG_LEVEL = def.UnsetTypeID
	}
	p.TypeMap.T_STACK_TRACE = typeCPStackTrace.ID
	p.TypeMap.T_CLASS_LOADER = typeCPClassLoader.ID

	typeStackFrame := p.TypeMap.NameMap["jdk.types.StackFrame"]

	if typeStackFrame == nil {
		return fmt.Errorf("missing \"jdk.types.StackFrame\"")
	}
	p.TypeMap.T_STACK_FRAME = typeStackFrame.ID

	p.bindFrameType = types2.NewBindFrameType(typeCPFrameType, &p.TypeMap)
	p.bindThreadState = types2.NewBindThreadState(typeCPThreadState, &p.TypeMap)
	p.bindThread = types2.NewBindThread(typeCPThread, &p.TypeMap)
	p.bindClass = types2.NewBindClass(typeCPClass, &p.TypeMap)
	p.bindMethod = types2.NewBindMethod(typeCPMethod, &p.TypeMap)
	p.bindPackage = types2.NewBindPackage(typeCPPackage, &p.TypeMap)
	p.bindSymbol = types2.NewBindSymbol(typeCPSymbol, &p.TypeMap)
	if typeCPLogLevel != nil {
		p.bindLogLevel = types2.NewBindLogLevel(typeCPLogLevel, &p.TypeMap)
	} else {
		p.bindLogLevel = nil
	}
	p.bindStackTrace = types2.NewBindStackTrace(typeCPStackTrace, &p.TypeMap)
	p.bindStackFrame = types2.NewBindStackFrame(typeStackFrame, &p.TypeMap)
	p.bindString = types2.NewBindString(tstring, &p.TypeMap)

	typeExecutionSample := p.TypeMap.NameMap["jdk.ExecutionSample"]
	typeWallClockSample := p.TypeMap.NameMap["profiler.WallClockSample"]
	typeAllocInNewTLAB := p.TypeMap.NameMap["jdk.ObjectAllocationInNewTLAB"]
	typeALlocOutsideTLAB := p.TypeMap.NameMap["jdk.ObjectAllocationOutsideTLAB"]
	typeAllocSample := p.TypeMap.NameMap["jdk.ObjectAllocationSample"]
	typeMonitorEnter := p.TypeMap.NameMap["jdk.JavaMonitorEnter"]
	typeThreadPark := p.TypeMap.NameMap["jdk.ThreadPark"]
	typeLiveObject := p.TypeMap.NameMap["profiler.LiveObject"]
	typeActiveSetting := p.TypeMap.NameMap["jdk.ActiveSetting"]

	typeMalloc := p.TypeMap.NameMap["profiler.Malloc"]
	typeFree := p.TypeMap.NameMap["profiler.Free"]

	initEventType(typeExecutionSample, &p.TypeMap.T_EXECUTION_SAMPLE, &p.bindExecutionSample, types2.NewBindExecutionSample, &p.TypeMap)
	initEventType(typeWallClockSample, &p.TypeMap.T_WALL_CLOCK_SAMPLE, &p.bindWallClockSample, types2.NewBindWallClockSample, &p.TypeMap)
	initEventType(typeMalloc, &p.TypeMap.T_MALLOC, &p.bindMalloc, types2.NewBindMalloc, &p.TypeMap)
	initEventType(typeFree, &p.TypeMap.T_FREE, &p.bindFree, types2.NewBindFree, &p.TypeMap)
	initEventType(typeAllocInNewTLAB, &p.TypeMap.T_ALLOC_IN_NEW_TLAB, &p.bindAllocInNewTLAB, types2.NewBindObjectAllocationInNewTLAB, &p.TypeMap)
	initEventType(typeALlocOutsideTLAB, &p.TypeMap.T_ALLOC_OUTSIDE_TLAB, &p.bindAllocOutsideTLAB, types2.NewBindObjectAllocationOutsideTLAB, &p.TypeMap)
	initEventType(typeAllocSample, &p.TypeMap.T_ALLOC_SAMPLE, &p.bindAllocSample, types2.NewBindObjectAllocationSample, &p.TypeMap)
	initEventType(typeMonitorEnter, &p.TypeMap.T_MONITOR_ENTER, &p.bindMonitorEnter, types2.NewBindJavaMonitorEnter, &p.TypeMap)
	initEventType(typeThreadPark, &p.TypeMap.T_THREAD_PARK, &p.bindThreadPark, types2.NewBindThreadPark, &p.TypeMap)
	initEventType(typeLiveObject, &p.TypeMap.T_LIVE_OBJECT, &p.bindLiveObject, types2.NewBindLiveObject, &p.TypeMap)
	initEventType(typeActiveSetting, &p.TypeMap.T_ACTIVE_SETTING, &p.bindActiveSetting, types2.NewBindActiveSetting, &p.TypeMap)

	p.FrameTypes.Reset()
	p.ThreadStates.Reset()
	p.Threads.Reset()
	p.Classes.Reset()
	p.Methods.Reset()
	p.Packages.Reset()
	p.Symbols.Reset()
	p.LogLevels.Reset()
	p.Stacktrace.Reset()
	p.Strings.Reset()
	return nil
}

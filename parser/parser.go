package parser

import (
	"fmt"
	"io"
	"unsafe"

	types2 "github.com/grafana/jfr-parser/parser/types"
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

	TypeMap types2.TypeMap
}

func NewParser(buf []byte, options Options) *Parser {
	p := &Parser{
		options: options,
		buf:     buf,
	}
	// Required cpool types
	p.TypeMap.T_STRING = types2.TypeBinding[types2.BindString]{Name: "java.lang.String", Factory: types2.NewBindString, Required: true}
	p.TypeMap.T_FRAME_TYPE = types2.TypeBinding[types2.BindFrameType]{Name: "jdk.types.FrameType", Factory: types2.NewBindFrameType, Required: true}
	p.TypeMap.T_THREAD_STATE = types2.TypeBinding[types2.BindThreadState]{Name: "jdk.types.ThreadState", Factory: types2.NewBindThreadState, Required: true}
	p.TypeMap.T_THREAD = types2.TypeBinding[types2.BindThread]{Name: "java.lang.Thread", Factory: types2.NewBindThread, Required: true}
	p.TypeMap.T_CLASS = types2.TypeBinding[types2.BindClass]{Name: "java.lang.Class", Factory: types2.NewBindClass, Required: true}
	p.TypeMap.T_METHOD = types2.TypeBinding[types2.BindMethod]{Name: "jdk.types.Method", Factory: types2.NewBindMethod, Required: true}
	p.TypeMap.T_PACKAGE = types2.TypeBinding[types2.BindPackage]{Name: "jdk.types.Package", Factory: types2.NewBindPackage, Required: true}
	p.TypeMap.T_SYMBOL = types2.TypeBinding[types2.BindSymbol]{Name: "jdk.types.Symbol", Factory: types2.NewBindSymbol, Required: true}
	p.TypeMap.T_STACK_TRACE = types2.TypeBinding[types2.BindStackTrace]{Name: "jdk.types.StackTrace", Factory: types2.NewBindStackTrace, Required: true}
	p.TypeMap.T_STACK_FRAME = types2.TypeBinding[types2.BindStackFrame]{Name: "jdk.types.StackFrame", Factory: types2.NewBindStackFrame, Required: true}
	p.TypeMap.T_CLASS_LOADER = types2.TypeBinding[types2.BindClassLoader]{Name: "jdk.types.ClassLoader", Factory: types2.NewBindClassLoader, Required: true}
	// Optional cpool type
	p.TypeMap.T_LOG_LEVEL = types2.TypeBinding[types2.BindLogLevel]{Name: "profiler.types.LogLevel", Factory: types2.NewBindLogLevel}
	// Optional event types
	p.TypeMap.T_EXECUTION_SAMPLE = types2.TypeBinding[types2.BindExecutionSample]{Name: "jdk.ExecutionSample", Factory: types2.NewBindExecutionSample}
	p.TypeMap.T_WALL_CLOCK_SAMPLE = types2.TypeBinding[types2.BindWallClockSample]{Name: "profiler.WallClockSample", Factory: types2.NewBindWallClockSample}
	p.TypeMap.T_MALLOC = types2.TypeBinding[types2.BindMalloc]{Name: "profiler.Malloc", Factory: types2.NewBindMalloc}
	p.TypeMap.T_FREE = types2.TypeBinding[types2.BindFree]{Name: "profiler.Free", Factory: types2.NewBindFree}
	p.TypeMap.T_ALLOC_IN_NEW_TLAB = types2.TypeBinding[types2.BindObjectAllocationInNewTLAB]{Name: "jdk.ObjectAllocationInNewTLAB", Factory: types2.NewBindObjectAllocationInNewTLAB}
	p.TypeMap.T_ALLOC_OUTSIDE_TLAB = types2.TypeBinding[types2.BindObjectAllocationOutsideTLAB]{Name: "jdk.ObjectAllocationOutsideTLAB", Factory: types2.NewBindObjectAllocationOutsideTLAB}
	p.TypeMap.T_ALLOC_SAMPLE = types2.TypeBinding[types2.BindObjectAllocationSample]{Name: "jdk.ObjectAllocationSample", Factory: types2.NewBindObjectAllocationSample}
	p.TypeMap.T_MONITOR_ENTER = types2.TypeBinding[types2.BindJavaMonitorEnter]{Name: "jdk.JavaMonitorEnter", Factory: types2.NewBindJavaMonitorEnter}
	p.TypeMap.T_THREAD_PARK = types2.TypeBinding[types2.BindThreadPark]{Name: "jdk.ThreadPark", Factory: types2.NewBindThreadPark}
	p.TypeMap.T_LIVE_OBJECT = types2.TypeBinding[types2.BindLiveObject]{Name: "profiler.LiveObject", Factory: types2.NewBindLiveObject}
	p.TypeMap.T_ACTIVE_SETTING = types2.TypeBinding[types2.BindActiveSetting]{Name: "jdk.ActiveSetting", Factory: types2.NewBindActiveSetting}
	return p
}

func (p *Parser) ParseEvent() (types2.TypeID, error) {
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
			return 0, types2.ErrIntOverflow
		}
		if uint64(p.chunkEnd-pp) < size {
			return 0, fmt.Errorf("invalid event size %d at position %d", size, pp)
		}
		typ, err := p.varLong()
		if err != nil {
			return 0, err
		}
		_ = size

		ttyp := types2.TypeID(typ)
		switch ttyp {
		case types2.UnsetTypeID:
			return types2.UnsetTypeID, fmt.Errorf("invalid event type %d at position %d", typ, pp)
		case p.TypeMap.T_EXECUTION_SAMPLE.TypeID:
			_, err = p.ExecutionSample.Parse(p.buf[p.pos:], p.TypeMap.T_EXECUTION_SAMPLE.Bind, &p.TypeMap)
		case p.TypeMap.T_WALL_CLOCK_SAMPLE.TypeID:
			_, err = p.WallClockSample.Parse(p.buf[p.pos:], p.TypeMap.T_WALL_CLOCK_SAMPLE.Bind, &p.TypeMap)
		case p.TypeMap.T_MALLOC.TypeID:
			_, err = p.Malloc.Parse(p.buf[p.pos:], p.TypeMap.T_MALLOC.Bind, &p.TypeMap)
		case p.TypeMap.T_FREE.TypeID:
			_, err = p.Free.Parse(p.buf[p.pos:], p.TypeMap.T_FREE.Bind, &p.TypeMap)
		case p.TypeMap.T_ALLOC_IN_NEW_TLAB.TypeID:
			_, err = p.ObjectAllocationInNewTLAB.Parse(p.buf[p.pos:], p.TypeMap.T_ALLOC_IN_NEW_TLAB.Bind, &p.TypeMap)
		case p.TypeMap.T_ALLOC_OUTSIDE_TLAB.TypeID:
			_, err = p.ObjectAllocationOutsideTLAB.Parse(p.buf[p.pos:], p.TypeMap.T_ALLOC_OUTSIDE_TLAB.Bind, &p.TypeMap)
		case p.TypeMap.T_ALLOC_SAMPLE.TypeID:
			_, err = p.ObjectAllocationSample.Parse(p.buf[p.pos:], p.TypeMap.T_ALLOC_SAMPLE.Bind, &p.TypeMap)
		case p.TypeMap.T_LIVE_OBJECT.TypeID:
			_, err = p.LiveObject.Parse(p.buf[p.pos:], p.TypeMap.T_LIVE_OBJECT.Bind, &p.TypeMap)
		case p.TypeMap.T_MONITOR_ENTER.TypeID:
			_, err = p.JavaMonitorEnter.Parse(p.buf[p.pos:], p.TypeMap.T_MONITOR_ENTER.Bind, &p.TypeMap)
		case p.TypeMap.T_THREAD_PARK.TypeID:
			_, err = p.ThreadPark.Parse(p.buf[p.pos:], p.TypeMap.T_THREAD_PARK.Bind, &p.TypeMap)
		case p.TypeMap.T_ACTIVE_SETTING.TypeID:
			_, err = p.ActiveSetting.Parse(p.buf[p.pos:], p.TypeMap.T_ACTIVE_SETTING.Bind, &p.TypeMap)
		default:
			p.pos = pp + int(size)
			continue
		}
		if err != nil {
			return types2.UnsetTypeID, err
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
			return 0, types2.ErrIntOverflow
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
		return "", types2.ErrIntOverflow
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
		return nil, types2.ErrIntOverflow
	}
	if p.pos+int(l) > len(p.buf) {
		return nil, io.ErrUnexpectedEOF
	}
	bs := p.buf[p.pos : p.pos+int(l)]
	p.pos += int(l)
	return bs, nil
}

func (p *Parser) checkTypes() error {
	// Resolve primitive types (no bind, just TypeID)
	primitives := []struct {
		name   string
		typeID *types2.TypeID
	}{
		{"int", &p.TypeMap.T_INT},
		{"long", &p.TypeMap.T_LONG},
		{"short", &p.TypeMap.T_SHORT},
		{"float", &p.TypeMap.T_FLOAT},
		{"boolean", &p.TypeMap.T_BOOLEAN},
	}
	for _, prim := range primitives {
		cls := p.TypeMap.NameMap[prim.name]
		if cls == nil {
			return fmt.Errorf("missing %q", prim.name)
		}
		*prim.typeID = cls.ID
	}

	// Resolve all type bindings (required ones error if absent)
	for _, b := range p.TypeMap.Bindings() {
		if err := b.Resolve(&p.TypeMap); err != nil {
			return err
		}
	}

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

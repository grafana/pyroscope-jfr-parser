package parser

import (
	"fmt"
	"io"
	"strings"
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

	DebugFile    io.Writer
	DebugTraceAll bool
	debugFoundRemaining bool
}

func NewParser(buf []byte, options Options) *Parser {
	p := &Parser{
		options: options,
		buf:     buf,
	}
	return p
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
		eventEnd := pp + int(size)
		if int(size) < 0 || eventEnd < pp || eventEnd > p.chunkEnd {
			err := fmt.Errorf("invalid event size %d at position %d", size, pp)
			if p.DebugFile != nil {
				fmt.Fprintf(p.DebugFile, "pos=%d size=%d err=%v\n", pp, size, err)
			}
			return 0, err
		}
		typ, err := p.varLong()
		if err != nil {
			return 0, err
		}

		ttyp := def.TypeID(typ)
		if p.DebugFile != nil && p.DebugTraceAll {
			name := fmt.Sprintf("%d", ttyp)
			if cls, ok := p.TypeMap.IDMap[ttyp]; ok {
				name = cls.Name
			}
			fmt.Fprintf(p.DebugFile, "pos=%d size=%d type=%s\n", pp, size, name)
		}
		switch ttyp {
		case p.TypeMap.T_EXECUTION_SAMPLE:
			if p.bindExecutionSample == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.ExecutionSample.Parse(p.buf[p.pos:], p.bindExecutionSample, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		case p.TypeMap.T_WALL_CLOCK_SAMPLE:
			if p.bindWallClockSample == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.WallClockSample.Parse(p.buf[p.pos:], p.bindWallClockSample, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		case p.TypeMap.T_MALLOC:
			if p.bindMalloc == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.Malloc.Parse(p.buf[p.pos:], p.bindMalloc, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		case p.TypeMap.T_FREE:
			if p.bindFree == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.Free.Parse(p.buf[p.pos:], p.bindFree, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		case p.TypeMap.T_ALLOC_IN_NEW_TLAB:
			if p.bindAllocInNewTLAB == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.ObjectAllocationInNewTLAB.Parse(p.buf[p.pos:], p.bindAllocInNewTLAB, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		case p.TypeMap.T_ALLOC_OUTSIDE_TLAB:
			if p.bindAllocOutsideTLAB == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.ObjectAllocationOutsideTLAB.Parse(p.buf[p.pos:], p.bindAllocOutsideTLAB, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		case p.TypeMap.T_ALLOC_SAMPLE:
			if p.bindAllocSample == nil {
				p.pos = eventEnd // skip
			}
			_, err := p.ObjectAllocationSample.Parse(p.buf[p.pos:], p.bindAllocSample, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		case p.TypeMap.T_LIVE_OBJECT:
			if p.bindLiveObject == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.LiveObject.Parse(p.buf[p.pos:], p.bindLiveObject, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		case p.TypeMap.T_MONITOR_ENTER:
			if p.bindMonitorEnter == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.JavaMonitorEnter.Parse(p.buf[p.pos:], p.bindMonitorEnter, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		case p.TypeMap.T_THREAD_PARK:
			if p.bindThreadPark == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.ThreadPark.Parse(p.buf[p.pos:], p.bindThreadPark, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil

		case p.TypeMap.T_ACTIVE_SETTING:
			if p.bindActiveSetting == nil {
				p.pos = eventEnd // skip
				continue
			}
			_, err := p.ActiveSetting.Parse(p.buf[p.pos:], p.bindActiveSetting, &p.TypeMap)
			if err != nil {
				return 0, err
			}
			p.pos = eventEnd
			return ttyp, nil
		default:
			if p.DebugFile != nil && (p.DebugTraceAll || !p.debugFoundRemaining) {
				p.reflectDebugDump(ttyp, pp, size, eventEnd)
			}
			p.pos = eventEnd
		}
	}
}

func (p *Parser) reflectDebugDump(ttyp def.TypeID, pp int, size uint64, eventEnd int) {
	cls, ok := p.TypeMap.IDMap[ttyp]
	if !ok {
		return
	}
	switch cls.Name {
	case "jdk.NetworkUtilization", "jdk.DataLoss",
		"jdk.ThreadSleep", "jdk.JavaExceptionThrow", "jdk.FileWrite",
		"jdk.InitialSystemProperty", "jdk.MetaspaceChunkFreeListSummary",
		"jdk.OldGarbageCollection":
	default:
		return
	}
	savedPos := p.pos
	defer func() { p.pos = savedPos }()

	var buf strings.Builder
	fmt.Fprintf(&buf, "  skipping pos=%d size=%d type=%s\n", pp, size, cls.Name)
	fmt.Fprintf(&buf, "    == %s fields ==\n", cls.Name)
	for _, f := range cls.Fields {
		if p.pos >= eventEnd {
			fmt.Fprintf(&buf, "    %s: <past event end>\n", f.Name)
			break
		}
		fieldType, _ := p.TypeMap.IDMap[f.Type]
		typeName := ""
		if fieldType != nil {
			typeName = fieldType.Name
		}
		if f.ConstantPool {
			v, err := p.varLong()
			if err != nil {
				fmt.Fprintf(&buf, "    %s: <err: %v>\n", f.Name, err)
				break
			}
			fmt.Fprintf(&buf, "    %s: ref=%d (type=%s, cp=true)\n", f.Name, v, typeName)
			if typeName == "java.lang.Class" {
				cls := p.GetClass(types2.ClassRef(v))
				if cls != nil {
					clsSym := p.GetSymbol(cls.Name)
					if clsSym != nil {
						fmt.Fprintf(&buf, "      -> %s\n", clsSym.String)
					}
				}
			}
			if typeName == "jdk.types.StackTrace" {
				st := p.GetStacktrace(types2.StackTraceRef(v))
				if st != nil {
					for i, frame := range st.Frames {
						m := p.GetMethod(frame.Method)
						if m == nil {
							fmt.Fprintf(&buf, "      [%d] method=%d line=%d\n", i, frame.Method, frame.LineNumber)
							continue
						}
						cls := p.GetClass(types2.ClassRef(m.Type))
						sym := p.GetSymbol(m.Name)
						clsName := ""
						if cls != nil {
							clsSym := p.GetSymbol(cls.Name)
							if clsSym != nil {
								clsName = clsSym.String
							}
						}
						methName := ""
						if sym != nil {
							methName = sym.String
						}
						fmt.Fprintf(&buf, "      [%d] %s.%s:%d\n", i, clsName, methName, frame.LineNumber)
					}
				}
			}
		} else if typeName == "long" || typeName == "int" || typeName == "short" || typeName == "byte" || typeName == "boolean" || typeName == "float" || typeName == "double" || typeName == "char" {
			v, err := p.varLong()
			if err != nil {
				fmt.Fprintf(&buf, "    %s: <err: %v>\n", f.Name, err)
				break
			}
			fmt.Fprintf(&buf, "    %s: %d (type=%s)\n", f.Name, v, typeName)
		} else if typeName == "java.lang.String" {
			s, err := p.string()
			if err != nil {
				fmt.Fprintf(&buf, "    %s: <err: %v>\n", f.Name, err)
				break
			}
			fmt.Fprintf(&buf, "    %s: %q (type=string)\n", f.Name, s)
		} else {
			fmt.Fprintf(&buf, "    %s: <unknown type %s>\n", f.Name, typeName)
			break
		}
	}
	remaining := eventEnd - p.pos
	fmt.Fprintf(&buf, "    pos after fields: %d, eventEnd: %d, remaining: %d\n", p.pos, eventEnd, remaining)
	if remaining > 0 {
		end := eventEnd
		if end > len(p.buf) {
			end = len(p.buf)
		}
		fmt.Fprintf(&buf, "    hexdump: %x\n", p.buf[p.pos:end])
		embSize, err := p.varLong()
		if err == nil {
			embType, err := p.varLong()
			if err == nil {
				name := fmt.Sprintf("%d", def.TypeID(embType))
				if c, ok := p.TypeMap.IDMap[def.TypeID(embType)]; ok {
					name = c.Name
				}
				fmt.Fprintf(&buf, "    embedded? size=%d type=%s\n", embSize, name)
			}
		}
	}

	if p.DebugTraceAll || (remaining > 0 && !p.debugFoundRemaining) {
		if remaining > 0 {
			p.debugFoundRemaining = true
		}
		fmt.Fprint(p.DebugFile, buf.String())
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
	if pos < len(p.buf) {
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
	switch b {
	case 0:
		return "", nil //todo this should be nil
	case 1:
		return "", nil
	case 2:
		// Constant pool reference — read the key and resolve the string.
		key, err := p.varLong()
		if err != nil {
			return "", err
		}
		idx, ok := p.Strings.IDMap[types2.StringRef(key)]
		if !ok {
			return "", nil
		}
		return p.Strings.String[idx].String, nil
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
		p.TypeMap.T_LOG_LEVEL = -1
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

	if typeExecutionSample != nil {
		p.TypeMap.T_EXECUTION_SAMPLE = typeExecutionSample.ID
		p.bindExecutionSample = types2.NewBindExecutionSample(typeExecutionSample, &p.TypeMap)
	} else {
		p.TypeMap.T_EXECUTION_SAMPLE = -1
		p.bindExecutionSample = nil
	}
	if typeWallClockSample != nil {
		p.TypeMap.T_WALL_CLOCK_SAMPLE = typeWallClockSample.ID
		p.bindWallClockSample = types2.NewBindWallClockSample(typeWallClockSample, &p.TypeMap)
	} else {
		p.TypeMap.T_WALL_CLOCK_SAMPLE = -1
		p.bindWallClockSample = nil
	}
	if typeMalloc != nil {
		p.TypeMap.T_MALLOC = typeMalloc.ID
		p.bindMalloc = types2.NewBindMalloc(typeMalloc, &p.TypeMap)
	} else {
		p.TypeMap.T_MALLOC = -1
		p.bindMalloc = nil
	}

	if typeFree != nil {
		p.TypeMap.T_FREE = typeFree.ID
		p.bindFree = types2.NewBindFree(typeFree, &p.TypeMap)
	} else {
		p.TypeMap.T_FREE = -1
		p.bindFree = nil
	}

	if typeAllocInNewTLAB != nil {
		p.TypeMap.T_ALLOC_IN_NEW_TLAB = typeAllocInNewTLAB.ID
		p.bindAllocInNewTLAB = types2.NewBindObjectAllocationInNewTLAB(typeAllocInNewTLAB, &p.TypeMap)
	} else {
		p.TypeMap.T_ALLOC_IN_NEW_TLAB = -1
		p.bindAllocInNewTLAB = nil
	}

	if typeALlocOutsideTLAB != nil {
		p.TypeMap.T_ALLOC_OUTSIDE_TLAB = typeALlocOutsideTLAB.ID
		p.bindAllocOutsideTLAB = types2.NewBindObjectAllocationOutsideTLAB(typeALlocOutsideTLAB, &p.TypeMap)
	} else {
		p.TypeMap.T_ALLOC_OUTSIDE_TLAB = -1
		p.bindAllocOutsideTLAB = nil
	}
	if typeAllocSample != nil {
		p.TypeMap.T_ALLOC_SAMPLE = typeAllocSample.ID
		p.bindAllocSample = types2.NewBindObjectAllocationSample(typeAllocSample, &p.TypeMap)
	} else {
		p.TypeMap.T_ALLOC_SAMPLE = -1
		p.bindAllocSample = nil
	}
	if typeMonitorEnter != nil {
		p.TypeMap.T_MONITOR_ENTER = typeMonitorEnter.ID
		p.bindMonitorEnter = types2.NewBindJavaMonitorEnter(typeMonitorEnter, &p.TypeMap)
	} else {
		p.TypeMap.T_MONITOR_ENTER = -1
		p.bindMonitorEnter = nil
	}

	if typeThreadPark != nil {
		p.TypeMap.T_THREAD_PARK = typeThreadPark.ID
		p.bindThreadPark = types2.NewBindThreadPark(typeThreadPark, &p.TypeMap)
	} else {
		p.TypeMap.T_THREAD_PARK = -1
		p.bindThreadPark = nil
	}

	if typeLiveObject != nil {
		p.TypeMap.T_LIVE_OBJECT = typeLiveObject.ID
		p.bindLiveObject = types2.NewBindLiveObject(typeLiveObject, &p.TypeMap)
	} else {
		p.TypeMap.T_LIVE_OBJECT = -1
		p.bindLiveObject = nil
	}

	if typeActiveSetting != nil {
		p.TypeMap.T_ACTIVE_SETTING = typeActiveSetting.ID
		p.bindActiveSetting = types2.NewBindActiveSetting(typeActiveSetting, &p.TypeMap)
	} else {
		p.TypeMap.T_ACTIVE_SETTING = -1
		p.bindActiveSetting = nil
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

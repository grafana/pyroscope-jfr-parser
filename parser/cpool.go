package parser

import (
	"fmt"

	types2 "github.com/grafana/jfr-parser/parser/types"
)

func (p *Parser) readConstantPool(pos int) error {
	for {
		if err := p.seek(pos); err != nil {
			return err
		}
		sz, err := p.varLong()
		if err != nil {
			return err
		}
		typ, err := p.varLong()
		if err != nil {
			return err
		}
		startTimeTicks, err := p.varLong()
		if err != nil {
			return err
		}
		duration, err := p.varLong()
		if err != nil {
			return err
		}
		delta, err := p.varLong()
		if err != nil {
			return err
		}
		typeMask, err := p.varInt() // boolean flush
		if err != nil {
			return err
		}
		n, err := p.varInt()
		if err != nil {
			return err
		}
		_ = startTimeTicks
		_ = duration
		_ = delta
		_ = sz
		_ = typeMask
		_ = typ

		id := int(int64(delta))

		for i := 0; i < int(n); i++ {
			typ, err := p.varLong()
			if err != nil {
				return err
			}
			c := p.TypeMap.IDMap[types2.TypeID(typ)]
			if c == nil {
				return fmt.Errorf("unknown type %d", types2.TypeID(typ))
			}
			err = p.readConstants(c)
			if err != nil {
				return fmt.Errorf("error reading %+v %w", c, err)
			}
		}
		if delta == 0 {
			break
		} else {
			pos += id
			if pos <= 0 {
				break
			}
		}
	}
	return nil
}

func (p *Parser) readConstants(c *types2.MetadataClass) error {
	switch c.Name {
	case "jdk.types.ChunkHeader":
		p.pos += chunkHeaderSize
		return nil
	case "jdk.types.FrameType":
		o, err := p.FrameTypes.Parse(p.buf[p.pos:], p.TypeMap.T_FRAME_TYPE.Bind, &p.TypeMap)
		p.pos += o
		return err
	case "jdk.types.ThreadState":
		o, err := p.ThreadStates.Parse(p.buf[p.pos:], p.TypeMap.T_THREAD_STATE.Bind, &p.TypeMap)
		p.pos += o
		return err
	case "java.lang.Thread":
		o, err := p.Threads.Parse(p.buf[p.pos:], p.TypeMap.T_THREAD.Bind, &p.TypeMap)
		p.pos += o
		return err
	case "java.lang.Class":
		o, err := p.Classes.Parse(p.buf[p.pos:], p.TypeMap.T_CLASS.Bind, &p.TypeMap)
		p.pos += o
		return err
	case "jdk.types.Method":
		o, err := p.Methods.Parse(p.buf[p.pos:], p.TypeMap.T_METHOD.Bind, &p.TypeMap)
		p.pos += o
		return err
	case "jdk.types.Package":
		o, err := p.Packages.Parse(p.buf[p.pos:], p.TypeMap.T_PACKAGE.Bind, &p.TypeMap)
		p.pos += o
		return err
	case "jdk.types.Symbol":
		o, err := p.Symbols.Parse(p.buf[p.pos:], p.TypeMap.T_SYMBOL.Bind, &p.TypeMap)
		p.pos += o
		return err
	case "profiler.types.LogLevel":
		if p.TypeMap.T_LOG_LEVEL.Bind == nil {
			return fmt.Errorf("no \"profiler.types.LogLevel\"")
		}
		o, err := p.LogLevels.Parse(p.buf[p.pos:], p.TypeMap.T_LOG_LEVEL.Bind, &p.TypeMap)
		p.pos += o
		return err
	case "jdk.types.StackTrace":
		o, err := p.Stacktrace.Parse(p.buf[p.pos:], p.TypeMap.T_STACK_TRACE.Bind, p.TypeMap.T_STACK_FRAME.Bind, &p.TypeMap)
		p.pos += o
		return err
	case "java.lang.String":
		o, err := p.Strings.Parse(p.buf[p.pos:], p.TypeMap.T_STRING.Bind, &p.TypeMap)
		p.pos += o
		return err
	default:
		b := types2.NewBindSkipConstantPool(c, &p.TypeMap)
		skipper := types2.SkipConstantPoolList{}
		o, err := skipper.Parse(p.buf[p.pos:], b, &p.TypeMap)
		p.pos += o
		return err
	}
}

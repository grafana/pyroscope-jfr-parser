package jfrwriter

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strconv"

	"github.com/grafana/jfr-parser/parser/types/def"
)

const chunkHeaderSize = 68
const chunkMagic = 0x464c5200
const chunkVersion = 0x00020000

type cpoolEntry struct {
	id   uint64
	data []byte
}

type rawEvent struct {
	typeID def.TypeID
	data   []byte
}

// Builder accumulates metadata classes, constant pool entries, and events,
// then serializes them into a valid JFR chunk.
type Builder struct {
	classes    []def.Class
	classIndex map[string]int // name -> index into classes

	cpoolEntries map[def.TypeID][]cpoolEntry
	events       []rawEvent

	startNanos     uint64
	durationNanos  uint64
	startTicks     uint64
	ticksPerSecond uint64
	features       uint32
}

// NewBuilder creates a new Builder with sensible defaults.
func NewBuilder() *Builder {
	return &Builder{
		classIndex:     make(map[string]int),
		cpoolEntries:   make(map[def.TypeID][]cpoolEntry),
		startNanos:     1000000000,
		durationNanos:  10000000,
		startTicks:     100,
		ticksPerSecond: 1000000000,
	}
}

// SetTiming configures the chunk timing fields.
func (b *Builder) SetTiming(startNanos, durationNanos, startTicks, ticksPerSecond uint64) *Builder {
	b.startNanos = startNanos
	b.durationNanos = durationNanos
	b.startTicks = startTicks
	b.ticksPerSecond = ticksPerSecond
	return b
}

// AddClasses registers class definitions for the metadata section.
// Duplicate names are replaced.
func (b *Builder) AddClasses(classes ...def.Class) *Builder {
	for _, c := range classes {
		if idx, ok := b.classIndex[c.Name]; ok {
			b.classes[idx] = c
		} else {
			b.classIndex[c.Name] = len(b.classes)
			b.classes = append(b.classes, c)
		}
	}
	return b
}

// AddCPoolEntry adds a constant pool entry for the given typeID.
// data contains the pre-serialized field values for the entry.
// id is the constant pool reference ID for this entry.
func (b *Builder) AddCPoolEntry(typeID def.TypeID, id uint64, data []byte) *Builder {
	b.cpoolEntries[typeID] = append(b.cpoolEntries[typeID], cpoolEntry{id: id, data: data})
	return b
}

// AddEvent adds an event with the given typeID.
// data contains the pre-serialized field values (size and typeID prefix are added by Build).
func (b *Builder) AddEvent(typeID def.TypeID, data []byte) *Builder {
	b.events = append(b.events, rawEvent{typeID: typeID, data: data})
	return b
}

// Build serializes the chunk into a complete JFR byte slice.
func (b *Builder) Build() ([]byte, error) {
	eventsData := b.buildEvents()
	cpoolData := b.buildConstantPool()
	metaData, err := b.buildMetadata()
	if err != nil {
		return nil, fmt.Errorf("building metadata: %w", err)
	}

	totalSize := chunkHeaderSize + len(eventsData) + len(cpoolData) + len(metaData)
	offsetCPool := chunkHeaderSize + len(eventsData)
	offsetMeta := chunkHeaderSize + len(eventsData) + len(cpoolData)

	buf := make([]byte, totalSize)
	// Chunk header (big-endian)
	binary.BigEndian.PutUint32(buf[0:], chunkMagic)
	binary.BigEndian.PutUint32(buf[4:], chunkVersion)
	binary.BigEndian.PutUint64(buf[8:], uint64(totalSize))
	binary.BigEndian.PutUint64(buf[16:], uint64(offsetCPool))
	binary.BigEndian.PutUint64(buf[24:], uint64(offsetMeta))
	binary.BigEndian.PutUint64(buf[32:], b.startNanos)
	binary.BigEndian.PutUint64(buf[40:], b.durationNanos)
	binary.BigEndian.PutUint64(buf[48:], b.startTicks)
	binary.BigEndian.PutUint64(buf[56:], b.ticksPerSecond)
	binary.BigEndian.PutUint32(buf[64:], b.features)

	copy(buf[chunkHeaderSize:], eventsData)
	copy(buf[offsetCPool:], cpoolData)
	copy(buf[offsetMeta:], metaData)

	return buf, nil
}

// buildEvents serializes all events.
// Each event: [size:varLong][typeID:varLong][data]
// size includes itself (from the parser: pp = pos before size, then pos = pp + size).
func (b *Builder) buildEvents() []byte {
	var w Writer
	for _, e := range b.events {
		innerLen := VarLongSize(uint64(e.typeID)) + len(e.data)
		totalSize := computeEventSize(innerLen)
		w.VarLong(uint64(totalSize))
		w.VarLong(uint64(e.typeID))
		w.Raw(e.data)
	}
	return w.Bytes()
}

// computeEventSize finds the total event size (including the size field itself).
// size = varLongSize(size) + innerLen, which is self-referential.
func computeEventSize(innerLen int) int {
	for candidateSize := innerLen + 1; candidateSize <= innerLen+10; candidateSize++ {
		if VarLongSize(uint64(candidateSize))+innerLen == candidateSize {
			return candidateSize
		}
	}
	return innerLen + 1
}

// buildConstantPool serializes the constant pool section.
func (b *Builder) buildConstantPool() []byte {
	// Serialize the content (everything after the size field).
	var content Writer
	content.VarLong(1) // type = T_CPOOL
	content.VarLong(0) // startTimeTicks
	content.VarLong(0) // duration
	content.VarLong(0) // delta = 0 (no chaining)
	content.VarInt(1)  // flush = true

	// Sort type IDs for deterministic output.
	typeIDs := make([]def.TypeID, 0, len(b.cpoolEntries))
	for tid := range b.cpoolEntries {
		typeIDs = append(typeIDs, tid)
	}
	sort.Slice(typeIDs, func(i, j int) bool { return typeIDs[i] < typeIDs[j] })

	content.VarInt(uint32(len(typeIDs)))
	for _, tid := range typeIDs {
		entries := b.cpoolEntries[tid]
		content.VarLong(uint64(tid))
		content.VarInt(uint32(len(entries)))
		for _, e := range entries {
			content.VarLong(e.id)
			content.Raw(e.data)
		}
	}

	// Wrap with size prefix. Size includes itself.
	contentBytes := content.Bytes()
	totalSize := computeEventSize(len(contentBytes))

	var w Writer
	w.VarLong(uint64(totalSize))
	w.Raw(contentBytes)
	return w.Bytes()
}

// metaElement represents a node in the metadata element tree.
type metaElement struct {
	name     string
	attrs    [][2]string
	children []metaElement
}

// buildMetadata serializes the metadata section.
func (b *Builder) buildMetadata() ([]byte, error) {
	// Build element tree
	tree := b.buildMetadataTree()

	// Collect all strings into a string table
	st := newStringTable()
	collectStrings(&tree, st)

	// Serialize string table
	var strBuf Writer
	for _, s := range st.strings() {
		strBuf.String(s)
	}

	// Serialize element tree
	var treeBuf Writer
	serializeElement(&treeBuf, &tree, st)

	// Assemble metadata section as an event-like structure:
	// [size:varLong][type=0:varLong][0:varLong][0:varLong][0:varLong]
	// [nStrings:varInt][strings...][element tree...]
	// size is self-inclusive (same as events and cpool).
	var content Writer
	content.VarLong(0) // type = T_METADATA (0)
	content.VarLong(0) // startTicks
	content.VarLong(0) // duration
	content.VarLong(0) // delta
	content.VarInt(uint32(len(st.strings())))
	content.Raw(strBuf.Bytes())
	content.Raw(treeBuf.Bytes())

	contentBytes := content.Bytes()
	totalSize := computeEventSize(len(contentBytes))

	var w Writer
	w.VarLong(uint64(totalSize))
	w.Raw(contentBytes)
	return w.Bytes(), nil
}

func (b *Builder) buildMetadataTree() metaElement {
	var classElements []metaElement
	for i := range b.classes {
		c := &b.classes[i]
		var fieldElements []metaElement
		for j := range c.Fields {
			f := &c.Fields[j]
			fieldAttrs := [][2]string{
				{"name", f.Name},
				{"class", strconv.FormatInt(int64(f.Type), 10)},
			}
			if f.ConstantPool {
				fieldAttrs = append(fieldAttrs, [2]string{"constantPool", "true"})
			}
			if f.Array {
				fieldAttrs = append(fieldAttrs, [2]string{"dimension", "1"})
			}
			fieldElements = append(fieldElements, metaElement{name: "field", attrs: fieldAttrs})
		}
		classAttrs := [][2]string{
			{"id", strconv.FormatInt(int64(c.ID), 10)},
			{"name", c.Name},
		}
		classElements = append(classElements, metaElement{name: "class", attrs: classAttrs, children: fieldElements})
	}

	return metaElement{
		name: "root",
		children: []metaElement{
			{name: "metadata", children: classElements},
			{name: "region"},
		},
	}
}

func collectStrings(e *metaElement, st *stringTable) {
	st.add(e.name)
	for _, attr := range e.attrs {
		st.add(attr[0])
		st.add(attr[1])
	}
	for i := range e.children {
		collectStrings(&e.children[i], st)
	}
}

func serializeElement(w *Writer, e *metaElement, st *stringTable) {
	w.VarInt(uint32(st.index(e.name)))
	w.VarInt(uint32(len(e.attrs)))
	for _, attr := range e.attrs {
		w.VarInt(uint32(st.index(attr[0])))
		w.VarInt(uint32(st.index(attr[1])))
	}
	w.VarInt(uint32(len(e.children)))
	for i := range e.children {
		serializeElement(w, &e.children[i], st)
	}
}

// stringTable maintains an ordered set of unique strings with index lookup.
type stringTable struct {
	strs    []string
	indices map[string]int
}

func newStringTable() *stringTable {
	return &stringTable{
		indices: make(map[string]int),
	}
}

func (st *stringTable) add(s string) {
	if _, ok := st.indices[s]; !ok {
		st.indices[s] = len(st.strs)
		st.strs = append(st.strs, s)
	}
}

func (st *stringTable) index(s string) int {
	return st.indices[s]
}

func (st *stringTable) strings() []string {
	return st.strs
}

package jfrwriter

import "bytes"

// Writer wraps a bytes.Buffer with JFR-specific encoding methods.
type Writer struct {
	buf bytes.Buffer
}

// VarInt encodes v as a JFR variable-length 32-bit integer (LEB128 variant).
func (w *Writer) VarInt(v uint32) {
	for v >= 0x80 {
		w.buf.WriteByte(byte(v&0x7F) | 0x80)
		v >>= 7
	}
	w.buf.WriteByte(byte(v))
}

// VarLong encodes v as a JFR variable-length 64-bit integer.
// Uses 7-bit groups with continuation bit. At shift==56, writes the full byte.
func (w *Writer) VarLong(v uint64) {
	for shift := uint(0); shift <= 56; shift += 7 {
		if shift == 56 {
			w.buf.WriteByte(byte(v >> 56))
			return
		}
		if v < 0x80 {
			w.buf.WriteByte(byte(v))
			return
		}
		w.buf.WriteByte(byte(v&0x7F) | 0x80)
		v >>= 7
	}
}

// String encodes s as a JFR string.
// Empty string -> type byte 1. Non-empty -> type byte 3 + varInt(len) + UTF-8 bytes.
func (w *Writer) String(s string) {
	if s == "" {
		w.buf.WriteByte(1) // empty string
		return
	}
	w.buf.WriteByte(3) // UTF-8
	w.VarInt(uint32(len(s)))
	w.buf.WriteString(s)
}

// Byte writes a single byte.
func (w *Writer) Byte(b byte) {
	w.buf.WriteByte(b)
}

// Bool writes a boolean as a single byte (0 or 1).
func (w *Writer) Bool(b bool) {
	if b {
		w.buf.WriteByte(1)
	} else {
		w.buf.WriteByte(0)
	}
}

// Raw writes raw bytes.
func (w *Writer) Raw(data []byte) {
	w.buf.Write(data)
}

// Bytes returns the accumulated bytes.
func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

// Len returns the current buffer length.
func (w *Writer) Len() int {
	return w.buf.Len()
}

// Reset clears the buffer.
func (w *Writer) Reset() {
	w.buf.Reset()
}

// VarIntSize returns the number of bytes needed to encode v as a varInt.
func VarIntSize(v uint32) int {
	n := 1
	for v >= 0x80 {
		n++
		v >>= 7
	}
	return n
}

// VarLongSize returns the number of bytes needed to encode v as a varLong.
func VarLongSize(v uint64) int {
	n := 1
	for shift := uint(0); shift <= 56; shift += 7 {
		if shift == 56 {
			return n
		}
		if v < 0x80 {
			return n
		}
		n++
		v >>= 7
	}
	return n
}

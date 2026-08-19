package colstream

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
)

type Writer struct {
	out io.Writer
}

func NewWriter(out io.Writer) *Writer { return &Writer{out: out} }

func (w *Writer) write(payload []byte) {
	if w == nil || w.out == nil {
		return
	}
	var hdr [4]byte
	binary.LittleEndian.PutUint32(hdr[:], uint32(len(payload)))
	_, _ = w.out.Write(hdr[:])
	_, _ = w.out.Write(payload)
}

type ColumnSet struct {
	writers []*Writer

	lastU64 []uint64
	written []bool

	lastBytes    [][]byte
	writtenBytes []bool
}

func NewColumnSet(n int) *ColumnSet {
	return &ColumnSet{
		writers:      make([]*Writer, n),
		lastU64:      make([]uint64, n),
		written:      make([]bool, n),
		lastBytes:    make([][]byte, n),
		writtenBytes: make([]bool, n),
	}
}

func (c *ColumnSet) Attach(col int, f *os.File) {
	if c == nil || col < 0 || col >= len(c.writers) {
		return
	}
	c.writers[col] = NewWriter(f)
}

func (c *ColumnSet) Len() int {
	if c == nil {
		return 0
	}
	return len(c.writers)
}

func (c *ColumnSet) changed(col int, bits uint64) bool {
	if c == nil || col < 0 || col >= len(c.writers) {
		return false
	}
	if c.written[col] && c.lastU64[col] == bits {
		return false
	}
	c.lastU64[col], c.written[col] = bits, true
	return true
}

func (c *ColumnSet) SetF32(col int, v float32) {
	bits := uint64(math.Float32bits(v))
	if !c.changed(col, bits) {
		return
	}
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(bits))
	c.writers[col].write(b[:])
}

func (c *ColumnSet) SetI32(col int, v int32) {
	if !c.changed(col, uint64(uint32(v))) {
		return
	}
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], uint32(v))
	c.writers[col].write(b[:])
}

func (c *ColumnSet) SetU32(col int, v uint32) {
	if !c.changed(col, uint64(v)) {
		return
	}
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	c.writers[col].write(b[:])
}

func (c *ColumnSet) SetU8(col int, v uint8) {
	if !c.changed(col, uint64(v)) {
		return
	}
	c.writers[col].write([]byte{v})
}

func (c *ColumnSet) SetBytes(col int, payload []byte) {
	if c == nil || col < 0 || col >= len(c.writers) {
		return
	}
	if c.writtenBytes[col] && bytes.Equal(c.lastBytes[col], payload) {
		return
	}
	c.lastBytes[col] = append(c.lastBytes[col][:0], payload...)
	c.writtenBytes[col] = true
	c.writers[col].write(payload)
}

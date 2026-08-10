// Package recread holds the byte-reading PRIMITIVES for one deframed input record.
//
// One job: turn bytes into numbers. Everything here is little-endian (matching the content
// buffer's encoding) and knows nothing about record kinds, entities, or attributes — what
// those bytes MEAN is nodes/Wiring's own decode files' job. A read past the end of the
// record is ErrShort, never a partial value the caller might mistake for data.
package recread

import (
	"encoding/binary"
	"errors"
	"math"
)

// ErrShort is returned by every read that would run past the end of the record.
var ErrShort = errors.New("input record truncated")

// Reader is a little-endian cursor over one deframed record body.
type Reader struct {
	b   []byte
	pos int
}

// NewReader returns a Reader over b, starting at pos (the caller has already consumed
// any leading tag byte(s) and passes the offset to resume from).
func NewReader(b []byte, pos int) *Reader {
	return &Reader{b: b, pos: pos}
}

func (r *Reader) U8() (byte, error) {
	if r.pos+1 > len(r.b) {
		return 0, ErrShort
	}
	v := r.b[r.pos]
	r.pos++
	return v, nil
}

func (r *Reader) I32() (int32, error) {
	if r.pos+4 > len(r.b) {
		return 0, ErrShort
	}
	v := int32(binary.LittleEndian.Uint32(r.b[r.pos:]))
	r.pos += 4
	return v, nil
}

// F32 reads a 4-byte float. The drop point rides as f32, not f64: it is a WORLD POSITION,
// the same precision every position on the content buffer already uses, and it is about to
// be rounded onto the node lattice anyway.
func (r *Reader) F32() (float32, error) {
	if r.pos+4 > len(r.b) {
		return 0, ErrShort
	}
	v := math.Float32frombits(binary.LittleEndian.Uint32(r.b[r.pos:]))
	r.pos += 4
	return v, nil
}

func (r *Reader) F64() (float64, error) {
	if r.pos+8 > len(r.b) {
		return 0, ErrShort
	}
	v := math.Float64frombits(binary.LittleEndian.Uint64(r.b[r.pos:]))
	r.pos += 8
	return v, nil
}

func (r *Reader) BoolByte() (bool, error) {
	v, err := r.U8()
	return v != 0, err
}

func EnumAt(list []string, i byte) string {
	if int(i) < len(list) {
		return list[i]
	}
	return ""
}

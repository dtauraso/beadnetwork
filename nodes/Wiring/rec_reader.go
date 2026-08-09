// rec_reader.go — the byte-reading PRIMITIVES for one deframed input record.
//
// One job: turn bytes into numbers. Everything here is little-endian (matching the content
// buffer's encoding) and knows nothing about record kinds, entities, or attributes — what
// those bytes MEAN is input_codec.go's job. A read past the end of the record is
// errShortRecord, never a partial value the caller might mistake for data.

package Wiring

import (
	"encoding/binary"
	"errors"
	"math"
)

var errShortRecord = errors.New("input record truncated")

// recReader is a little-endian cursor over one deframed record body.
type recReader struct {
	b   []byte
	pos int
}

func (r *recReader) u8() (byte, error) {
	if r.pos+1 > len(r.b) {
		return 0, errShortRecord
	}
	v := r.b[r.pos]
	r.pos++
	return v, nil
}

func (r *recReader) i32() (int32, error) {
	if r.pos+4 > len(r.b) {
		return 0, errShortRecord
	}
	v := int32(binary.LittleEndian.Uint32(r.b[r.pos:]))
	r.pos += 4
	return v, nil
}

// f32 reads a 4-byte float. The drop point rides as f32, not f64: it is a WORLD POSITION,
// the same precision every position on the content buffer already uses, and it is about to
// be rounded onto the node lattice anyway.
func (r *recReader) f32() (float32, error) {
	if r.pos+4 > len(r.b) {
		return 0, errShortRecord
	}
	v := math.Float32frombits(binary.LittleEndian.Uint32(r.b[r.pos:]))
	r.pos += 4
	return v, nil
}

func (r *recReader) f64() (float64, error) {
	if r.pos+8 > len(r.b) {
		return 0, errShortRecord
	}
	v := math.Float64frombits(binary.LittleEndian.Uint64(r.b[r.pos:]))
	r.pos += 8
	return v, nil
}

func (r *recReader) boolByte() (bool, error) {
	v, err := r.u8()
	return v != 0, err
}

func enumAt(list []string, i byte) string {
	if int(i) < len(list) {
		return list[i]
	}
	return ""
}

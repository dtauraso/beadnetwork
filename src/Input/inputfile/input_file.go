// Package inputfile is the current input, as one file.
//
// The bytes are the SAME record the stdin path carries — the encoder, the
// decoder and INPUT_LAYOUT_FINGERPRINT are unchanged, so the two sides cannot
// drift and no second format exists. What changes is that the reader looks when
// it wakes, instead of being handed work by a queue.
package inputfile

import (
	"os"
)

// Reader hands back a record only when the bytes on disk differ from the ones
// it last returned. Reading the same input twice would apply it twice, and a
// wheel delta applied twice is a zoom that keeps going.
type Reader struct {
	path string
	last []byte
}

func NewReader(inputPath string) *Reader { return &Reader{path: inputPath} }

func (r *Reader) Read() ([]byte, bool) {
	raw, err := os.ReadFile(r.path)
	if err != nil || len(raw) == 0 {
		return nil, false
	}
	if string(raw) == string(r.last) {
		return nil, false
	}
	r.last = append(r.last[:0], raw...)
	return raw, true
}

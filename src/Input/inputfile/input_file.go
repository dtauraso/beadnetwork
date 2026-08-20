package inputfile

import (
	"os"
)

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

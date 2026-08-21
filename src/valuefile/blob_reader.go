package valuefile

import (
	"encoding/binary"
	"math"
	"os"
)

type BlobReader struct {
	values map[string][]byte
}

func ReadBlob(path string, names []string) (*BlobReader, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	values := make(map[string][]byte, len(names))
	off := 0
	for _, name := range names {
		if off+4 > len(raw) {
			return nil, false
		}
		n := int(binary.LittleEndian.Uint32(raw[off:]))
		off += 4
		if off+n > len(raw) {
			return nil, false
		}
		values[name] = raw[off : off+n]
		off += n
	}
	return &BlobReader{values: values}, true
}

func (r *BlobReader) Bool(name string) (bool, bool) {
	v, ok := r.values[name]
	if !ok || len(v) < 1 {
		return false, false
	}
	return v[0] != 0, true
}

func (r *BlobReader) F64(name string) (float64, bool) {
	v, ok := r.values[name]
	if !ok || len(v) < 8 {
		return 0, false
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(v)), true
}

func (r *BlobReader) I64(name string) (int64, bool) {
	v, ok := r.values[name]
	if !ok || len(v) < 8 {
		return 0, false
	}
	return int64(binary.LittleEndian.Uint64(v)), true
}

func (r *BlobReader) Text(name string) (string, bool) {
	v, ok := r.values[name]
	if !ok {
		return "", false
	}
	return string(v), true
}

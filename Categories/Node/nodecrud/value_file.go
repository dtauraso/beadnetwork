package nodecrud

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

func WriteAtomic(path string, v any) error {
	out, err := encode(v)
	if err != nil {
		return err
	}
	return writeBytesAtomic(path, out)
}

const atomicWriteTmpSuffix = ".tmp"

func encode(v any) ([]byte, error) {
	b := make([]byte, 8)
	switch t := v.(type) {
	case bool:
		if t {
			return []byte{1}, nil
		}
		return []byte{0}, nil
	case int:
		binary.LittleEndian.PutUint64(b, uint64(int64(t)))
		return b, nil
	case int32:
		binary.LittleEndian.PutUint64(b, uint64(int64(t)))
		return b, nil
	case int64:
		binary.LittleEndian.PutUint64(b, uint64(t))
		return b, nil
	case float64:
		binary.LittleEndian.PutUint64(b, math.Float64bits(t))
		return b, nil
	case string:
		return []byte(t), nil
	}
	return nil, fmt.Errorf("valuefile: %T is not a leaf primitive — a file holds one bool, integer, float64 or string", v)
}

func writeBytesAtomic(path string, out []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + atomicWriteTmpSuffix
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

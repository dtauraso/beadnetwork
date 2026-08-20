package valuefile

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const atomicWriteTmpSuffix = ".tmp"

const Ext = ".bin"

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

func decode(raw []byte, dst any) error {
	switch p := dst.(type) {
	case *bool:
		if len(raw) != 1 {
			return fmt.Errorf("valuefile: bool leaf is %d bytes, want 1", len(raw))
		}
		*p = raw[0] != 0
		return nil
	case *int:
		v, err := readInt64(raw)
		*p = int(v)
		return err
	case *int32:
		v, err := readInt64(raw)
		*p = int32(v)
		return err
	case *int64:
		v, err := readInt64(raw)
		*p = v
		return err
	case *float64:
		if len(raw) != 8 {
			return fmt.Errorf("valuefile: float leaf is %d bytes, want 8", len(raw))
		}
		*p = math.Float64frombits(binary.LittleEndian.Uint64(raw))
		return nil
	case *string:
		*p = string(raw)
		return nil
	}
	return fmt.Errorf("valuefile: %T is not a leaf primitive pointer", dst)
}

func readInt64(raw []byte) (int64, error) {
	if len(raw) != 8 {
		return 0, fmt.Errorf("valuefile: integer leaf is %d bytes, want 8", len(raw))
	}
	return int64(binary.LittleEndian.Uint64(raw)), nil
}

func SafeTreePathComponent(s string) bool {
	if s == "" || s == "." || s == ".." {
		return false
	}
	return !strings.ContainsAny(s, `/\`)
}

func LogPersistErr(label, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s: persist %s: %v\n", label, path, err)
}

func ReadIfExists(path string, dst any) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	if err := decode(raw, dst); err != nil {
		LogPersistErr("valuefile: CORRUPT LEAF", path, err)
		return false
	}
	return true
}

func RemoveIfPresent(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func WriteAtomicIfChanged(path string, v any) error {
	out, err := encode(v)
	if err != nil {
		return err
	}
	if prev, err := os.ReadFile(path); err == nil && string(prev) == string(out) {
		return nil
	}
	return writeBytesAtomic(path, out)
}

func WriteAtomic(path string, v any) error {
	out, err := encode(v)
	if err != nil {
		return err
	}
	return writeBytesAtomic(path, out)
}

// WriteVecAtomicIfChanged writes several float64s as ONE file, so a reader sees
// all of them or none. Values that change together must arrive together: three
// components of a vector in three files have no atomicity across them, and a
// read landing between two writes returns a vector that never existed.
func WriteVecAtomicIfChanged(path string, vs ...float64) error {
	out := make([]byte, 0, len(vs)*8)
	for _, v := range vs {
		out = binary.LittleEndian.AppendUint64(out, math.Float64bits(v))
	}
	if prev, err := os.ReadFile(path); err == nil && string(prev) == string(out) {
		return nil
	}
	return writeBytesAtomic(path, out)
}

func ReadVecIfExists(path string, n int) ([]float64, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	if len(raw) != n*8 {
		LogPersistErr("valuefile: CORRUPT VEC", path, fmt.Errorf("vec leaf is %d bytes, want %d", len(raw), n*8))
		return nil, false
	}
	out := make([]float64, n)
	for i := range out {
		out[i] = math.Float64frombits(binary.LittleEndian.Uint64(raw[i*8:]))
	}
	return out, true
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

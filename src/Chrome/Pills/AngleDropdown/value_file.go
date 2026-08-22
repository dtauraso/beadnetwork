
package AngleDropdown

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

type BlobWriter struct {
	path	string
	names	[]string
	known	map[string]bool
	pending	map[string][]byte
	last	[]byte
}

func (w *BlobWriter) Begin() {
	w.pending = make(map[string][]byte, len(w.names))
}

func (w *BlobWriter) Bool(name string, v bool) {
	if v {
		w.U8(name, 1)
		return
	}
	w.U8(name, 0)
}

func (w *BlobWriter) Bytes(name string, b []byte) {
	w.put(name, b)
}

func (w *BlobWriter) F32(name string, v float32) {
	w.put(name, binary.LittleEndian.AppendUint32(w.pending[name], math.Float32bits(v)))
}

func (w *BlobWriter) F64(name string, v float64) {
	w.put(name, binary.LittleEndian.AppendUint64(w.pending[name], math.Float64bits(v)))
}

func (w *BlobWriter) Flush() error {
	var out []byte
	for _, name := range w.names {
		v := w.pending[name]
		out = binary.LittleEndian.AppendUint32(out, uint32(len(v)))
		out = append(out, v...)
	}
	w.pending = nil

	if bytes.Equal(w.last, out) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(w.path), 0o755); err != nil {
		return err
	}
	tmp := w.path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, w.path); err != nil {
		return err
	}
	w.last = out
	return nil
}

func (w *BlobWriter) I32(name string, v int32) {
	w.put(name, binary.LittleEndian.AppendUint32(w.pending[name], uint32(v)))
}

func (w *BlobWriter) I64(name string, v int64) {
	w.put(name, binary.LittleEndian.AppendUint64(w.pending[name], uint64(v)))
}

func (w *BlobWriter) Str(dataName, lenName, s string) {
	w.Text(dataName, s)
	w.U32(lenName, uint32(len(s)))
}

func (w *BlobWriter) Text(name string, s string) {
	w.put(name, append(w.pending[name], s...))
}

func (w *BlobWriter) U32(name string, v uint32) {
	w.put(name, binary.LittleEndian.AppendUint32(w.pending[name], v))
}

func (w *BlobWriter) U8(name string, v uint8) {
	w.put(name, append(w.pending[name], v))
}

func (w *BlobWriter) put(name string, b []byte) {
	if !w.known[name] {
		panic("blockFileWriter.set: no value named " + name + " is declared by this writer, so Go and the renderer disagree about what crosses the seam; add it to this concern's declared value names and run go generate ./...")
	}
	w.pending[name] = b
}

func LogPersistErr(label, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s: persist %s: %v\n", label, path, err)
}

func NewBlobWriter(path string, names []string) *BlobWriter {
	known := make(map[string]bool, len(names))
	for _, n := range names {
		known[n] = true
	}
	return &BlobWriter{path: path, names: names, known: known}
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

func WriteAtomic(path string, v any) error {
	out, err := encode(v)
	if err != nil {
		return err
	}
	return writeBytesAtomic(path, out)
}

const atomicWriteTmpSuffix = ".tmp"

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

func readInt64(raw []byte) (int64, error) {
	if len(raw) != 8 {
		return 0, fmt.Errorf("valuefile: integer leaf is %d bytes, want 8", len(raw))
	}
	return int64(binary.LittleEndian.Uint64(raw)), nil
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

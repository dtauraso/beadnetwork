package Flags

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
)

type BlobReader struct {
	values map[string][]byte
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

type BlobWriter struct {
	path    string
	names   []string
	known   map[string]bool
	pending map[string][]byte
	last    []byte
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

func NewBlobWriter(path string, names []string) *BlobWriter {
	known := make(map[string]bool, len(names))
	for _, n := range names {
		known[n] = true
	}
	return &BlobWriter{path: path, names: names, known: known}
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

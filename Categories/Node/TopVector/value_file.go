package TopVector

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
)

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

func (w *BlobWriter) F32(name string, v float32) {
	w.put(name, binary.LittleEndian.AppendUint32(w.pending[name], math.Float32bits(v)))
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

func (w *BlobWriter) U8(name string, v uint8) {
	w.put(name, append(w.pending[name], v))
}

func (w *BlobWriter) put(name string, b []byte) {
	if !w.known[name] {
		panic("TopVector.BlobWriter.put: no value named " + name + " is declared by this writer, so Go and the renderer disagree about what crosses the seam; add it to this concern's declared value names and run go generate ./...")
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

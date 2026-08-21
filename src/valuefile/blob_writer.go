package valuefile

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"strings"
)

func BlobFileName(name string) string {
	var b strings.Builder
	for i, r := range name {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r - 'A' + 'a')
			continue
		}
		b.WriteRune(r)
	}
	b.WriteString(".bin")
	return b.String()
}

type BlobWriter struct {
	dir     string
	names   []string
	known   map[string]bool
	pending map[string][]byte
	last    map[string][]byte
}

func NewBlobWriter(dir string, names []string) *BlobWriter {
	known := make(map[string]bool, len(names))
	for _, n := range names {
		known[n] = true
	}
	return &BlobWriter{dir: dir, names: names, known: known, last: map[string][]byte{}}
}

func (w *BlobWriter) Begin() {
	w.pending = make(map[string][]byte, len(w.names))
}

func (w *BlobWriter) put(name string, b []byte) {
	if !w.known[name] {
		panic("valuefile.BlobWriter: " + name + " is not one of this writer's declared values, so Go and the renderer disagree about what crosses; add it to the declaring list and run go generate ./...")
	}
	w.pending[name] = b
}

func (w *BlobWriter) F32(name string, v float32) {
	w.put(name, binary.LittleEndian.AppendUint32(w.pending[name], math.Float32bits(v)))
}

func (w *BlobWriter) U8(name string, v uint8) {
	w.put(name, append(w.pending[name], v))
}

func (w *BlobWriter) I32(name string, v int32) {
	w.put(name, binary.LittleEndian.AppendUint32(w.pending[name], uint32(v)))
}

func (w *BlobWriter) U32(name string, v uint32) {
	w.put(name, binary.LittleEndian.AppendUint32(w.pending[name], v))
}

func (w *BlobWriter) Bool(name string, v bool) {
	if v {
		w.U8(name, 1)
		return
	}
	w.U8(name, 0)
}

func (w *BlobWriter) Text(name string, s string) {
	w.put(name, append(w.pending[name], s...))
}

func (w *BlobWriter) Str(dataName, lenName, s string) {
	w.Text(dataName, s)
	w.U32(lenName, uint32(len(s)))
}

func (w *BlobWriter) Flush() error {
	made := false
	for _, name := range w.names {
		next := w.pending[name]
		if bytes.Equal(w.last[name], next) {
			continue
		}
		if !made {
			if err := os.MkdirAll(w.dir, 0o755); err != nil {
				return err
			}
			made = true
		}
		path := filepath.Join(w.dir, BlobFileName(name))
		tmp := path + ".tmp"
		if err := os.WriteFile(tmp, next, 0o644); err != nil {
			return err
		}
		if err := os.Rename(tmp, path); err != nil {
			return err
		}
		w.last[name] = append(w.last[name][:0:0], next...)
	}
	w.pending = nil
	return nil
}

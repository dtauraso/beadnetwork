package PolarRulesPanel

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
)

type ValueWriter struct {
	sceneRoot string
	pending   map[string][]byte
	last      map[string][]byte
}

func NewValueWriter(sceneRoot string) *ValueWriter {
	return &ValueWriter{sceneRoot: sceneRoot, last: map[string][]byte{}}
}

func (w *ValueWriter) Begin() {
	w.pending = make(map[string][]byte, len(PanelValueNames))
}

func (w *ValueWriter) F32(name string, v float32) {
	w.pending[name] = binary.LittleEndian.AppendUint32(w.pending[name], math.Float32bits(v))
}

func (w *ValueWriter) U8(name string, v uint8) {
	w.pending[name] = append(w.pending[name], v)
}

func (w *ValueWriter) I32(name string, v int32) {
	w.pending[name] = binary.LittleEndian.AppendUint32(w.pending[name], uint32(v))
}

func (w *ValueWriter) U32(name string, v uint32) {
	w.pending[name] = binary.LittleEndian.AppendUint32(w.pending[name], v)
}

func (w *ValueWriter) Bool(name string, v bool) {
	if v {
		w.U8(name, 1)
		return
	}
	w.U8(name, 0)
}

func (w *ValueWriter) Text(name string, s string) {
	w.pending[name] = append(w.pending[name], s...)
}

func (w *ValueWriter) Str(dataName, lenName, s string) {
	w.Text(dataName, s)
	w.U32(lenName, uint32(len(s)))
}

func (w *ValueWriter) Rect(xName, yName, wName, hName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
	w.F32(wName, r.W)
	w.F32(hName, r.H)
}

func (w *ValueWriter) Point(xName, yName string, r Rect) {
	w.F32(xName, r.X)
	w.F32(yName, r.Y)
}

func (w *ValueWriter) Flush() error {
	dir := filepath.Join(w.sceneRoot, filepath.FromSlash(ValueRelDir))
	made := false
	for _, name := range PanelValueNames {
		next := w.pending[name]
		if bytes.Equal(w.last[name], next) {
			continue
		}
		if !made {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			made = true
		}
		path := filepath.Join(dir, ValueFileName(name))
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

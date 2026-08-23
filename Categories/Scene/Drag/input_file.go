package Drag

import (
	"os"
	"path/filepath"
)

type inputSlot struct {
	path string
	last []byte
}

type InputDirReader struct {
	slots []inputSlot
}

func NewInputDirReader(inputDir string) *InputDirReader {
	slots := make([]inputSlot, 0, len(EventKinds))
	for _, kind := range EventKinds {
		path := filepath.Join(inputDir, kind+".bin")

		raw, err := os.ReadFile(path)
		if err != nil {
			raw = nil
		}
		slots = append(slots, inputSlot{path: path, last: raw})
	}
	return &InputDirReader{slots: slots}
}

func (r *InputDirReader) ReadAll() [][]byte {
	var out [][]byte
	for i := range r.slots {
		s := &r.slots[i]
		raw, err := os.ReadFile(s.path)
		if err != nil || len(raw) == 0 {
			continue
		}
		if string(raw) == string(s.last) {
			continue
		}
		s.last = append(s.last[:0], raw...)
		out = append(out, append([]byte(nil), raw...))
	}
	return out
}

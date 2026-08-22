package File

import (
	"os"
	"path/filepath"
)

type slot struct {
	path string
	last []byte
}

type Reader struct {
	slots []slot
}

func NewReader(inputDir string) *Reader {
	slots := make([]slot, 0, len(InEventKinds))
	for _, kind := range InEventKinds {
		path := filepath.Join(inputDir, kind+".bin")

		raw, err := os.ReadFile(path)
		if err != nil {
			raw = nil
		}
		slots = append(slots, slot{path: path, last: raw})
	}
	return &Reader{slots: slots}
}

func (r *Reader) ReadAll() [][]byte {
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

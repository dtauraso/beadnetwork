package inputfile

import (
	"os"
	"path/filepath"

	"github.com/dtauraso/wirefold/src/Input/inputcodec"
)

type slot struct {
	path string
	last []byte
}

type Reader struct {
	slots []slot
}

func NewReader(inputDir string) *Reader {
	slots := make([]slot, 0, len(inputcodec.EventKinds))
	for _, kind := range inputcodec.EventKinds {
		path := filepath.Join(inputDir, kind+".bin")

		// Seed last with what is already there. A slot holds the LAST event of
		// its kind, and it outlives the process: without this, every start reads
		// all seven slots as fresh and replays the previous session's final
		// press, motion, home, delete and key. Only what is written after we
		// start is input.
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

package Sheet

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

const stateRelDir = "view/chrome/sheet-state"

func (s *State) load(sceneRoot string) {
	s.dir = filepath.Join(sceneRoot, filepath.FromSlash(stateRelDir))
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		label := e.Name()
		if LoadScalar(s.file(label, "open"), 0, 0) == 1 {
			if s.Open == nil {
				s.Open = map[string]bool{}
			}
			s.Open[label] = true
		}
		if v := LoadScalar(s.file(label, "scroll"), 0, 0); v > 0 {
			s.SetScroll(label, float32(v))
		}
		if v := LoadScalar(s.file(label, "scroll-x"), 0, 0); v > 0 {
			s.SetScrollX(label, float32(v))
		}
		if v := LoadScalar(s.file(label, "height"), 0, 0); v > 0 {
			if s.Height == nil {
				s.Height = map[string]float32{}
			}
			s.Height[label] = float32(v)
		}
		if LoadScalar(s.file(label, "max"), 0, 0) == 1 {
			if s.Max == nil {
				s.Max = map[string]bool{}
			}
			s.Max[label] = true
		}
	}
}

func (s *State) file(label, name string) string {
	return filepath.Join(s.dir, label, name+".bin")
}

func (s *State) save(label, name string, v int32) {
	if s.dir == "" {
		return
	}
	if err := WriteScalar(s.file(label, name), v); err != nil {
		fmt.Fprintf(os.Stderr, "sheet_state_persist: %s %s: %v\n", label, name, err)
	}
}

func px(v float32) int32 { return int32(math.Round(float64(v))) }

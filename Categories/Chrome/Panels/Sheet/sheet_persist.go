package Sheet

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const stateRelDir = "view/chrome/sheet-state"

func (s *State) load(sceneRoot string) {
	s.dir = filepath.Join(sceneRoot, filepath.FromSlash(stateRelDir))
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}
	pos := map[string]int{}
	defer func() {
		for l := range pos {
			s.Order = append(s.Order, l)
		}
		slices.SortFunc(s.Order, func(a, b string) int { return pos[a] - pos[b] })
	}()
	for _, e := range entries {
		label := e.Name()
		if v := LoadScalar(s.file(label, "tab-pos"), -1, 0); v >= 0 {
			pos[label] = v
		}
		if LoadScalar(s.file(label, "open"), 0, 0) == 1 {
			if s.Open == nil {
				s.Open = map[string]bool{}
			}
			s.Open[label] = true
		}
		if v := LoadScalar(s.file(label, "scroll"), 0, 0); v > 0 {
			s.SetScroll(label, float32(v))
		}
		s.loadScrollX(label)
		if v := LoadScalar(s.file(label, popScrollFile), 0, 0); v > 0 {
			s.SetPopScroll(label, float32(v))
		}
		if v := LoadScalar(s.file(label, "height"), 0, 0); float32(v) >= 2*barH() {
			if s.Height == nil {
				s.Height = map[string]float32{}
			}
			s.Height[label] = float32(v)
		}
		if label == DockKey {
			s.Tab = int(LoadScalar(s.file(label, "tab"), 0, 0))
		}
		if LoadScalar(s.file(label, "folded"), 0, 0) == 1 {
			if s.Folded == nil {
				s.Folded = map[string]bool{}
			}
			s.Folded[label] = true
		}
	}
}

func (s *State) loadScrollX(label string) {
	files, _ := os.ReadDir(filepath.Join(s.dir, label))
	for _, f := range files {
		table, ok := strings.CutPrefix(strings.TrimSuffix(f.Name(), ".bin"), scrollXFile)
		if !ok {
			continue
		}
		if v := LoadScalar(s.file(label, scrollXFile+table), 0, 0); v > 0 {
			if s.ScrollX == nil {
				s.ScrollX = map[string]float32{}
			}
			s.ScrollX[slideKey(label, table)] = float32(v)
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

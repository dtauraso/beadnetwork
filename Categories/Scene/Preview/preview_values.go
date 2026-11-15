package Preview

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Vectors"
)

const ValueRelFile = "view/links-preview.bin"

var PreviewValueNames = []string{
	"hotSrc", "hotDst",
	"edgePick",
	"ghostX", "ghostY", "ghostZ",
	"edgeSrc", "edgeDst", "edgeState",
}

const (
	EdgeUnpicked uint8 = 1
	EdgePicked   uint8 = 2
	EdgeOpen     uint8 = 3
)

type EdgeMark struct {
	Src, Dst int32
	State    uint8
}

func ValueRelPath() string { return ValueRelFile }

type Preview struct {
	HotSrc, HotDst int32
	EdgePick       bool
	Ghosts         []Vectors.Vec3
	Edges          []EdgeMark
}

func None() Preview { return Preview{HotSrc: -1, HotDst: -1} }

type State struct{ w *BlobWriter }

func (s *State) Arm(sceneRoot string) {
	s.w = NewBlobWriter(filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile)), PreviewValueNames)
}

func (s *State) Write(p Preview) {
	w := s.w
	if w == nil {
		return
	}
	w.Begin()
	w.I32("hotSrc", p.HotSrc)
	w.I32("hotDst", p.HotDst)
	w.Bool("edgePick", p.EdgePick)
	for _, g := range p.Ghosts {
		w.F32("ghostX", float32(g.X))
		w.F32("ghostY", float32(g.Y))
		w.F32("ghostZ", float32(g.Z))
	}
	for _, e := range p.Edges {
		w.I32("edgeSrc", e.Src)
		w.I32("edgeDst", e.Dst)
		w.U8("edgeState", e.State)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "links_preview_values: %v\n", err)
	}
}

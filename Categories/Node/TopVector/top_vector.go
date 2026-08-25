package TopVector

import (
	"github.com/dtauraso/beadnetwork/Categories/Vectors/polarindex"
)

type Holder struct {
	delta polarindex.Offset
	set   bool

	runner *Runner
}

func (h *Holder) SetDelta(off polarindex.Offset) {
	h.delta, h.set = off, true
}

func (h *Holder) Delta() (polarindex.Offset, bool) { return h.delta, h.set }

func (h *Holder) ShiftBy(off polarindex.Offset) {
	h.delta = polarindex.Sum(h.delta, off)
}

func (h *Holder) TargetIndex(self polarindex.Index, sc polarindex.SceneConstants) polarindex.Index {
	return polarindex.Compose(self, h.delta, sc)
}

func (h *Holder) SetRunner(r *Runner) { h.runner = r }

func (h *Holder) Armed() bool { return h.runner != nil }

func (h *Holder) Runner() *Runner { return h.runner }

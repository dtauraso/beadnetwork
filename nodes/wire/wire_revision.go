package wire

import "github.com/dtauraso/wirefold/nodes/spatial"

type WireRevision struct {
	Steps int
	Seg   spatial.WireSegment
}

type revisionSlot struct {
	ch chan WireRevision
}

func newRevisionSlot() revisionSlot {
	return revisionSlot{ch: make(chan WireRevision, 1)}
}

func (s *revisionSlot) post(rev WireRevision) {
	if s.ch == nil {
		return
	}
	select {
	case <-s.ch:
	default:
	}
	select {
	case s.ch <- rev:
	default:
	}
}

func (s *revisionSlot) take() (WireRevision, bool) {
	if s.ch == nil {
		return WireRevision{}, false
	}
	select {
	case rev := <-s.ch:
		return rev, true
	default:
		return WireRevision{}, false
	}
}

func (pw *PacedWire) PostGeom(steps int, seg spatial.WireSegment) {
	if pw == nil {
		return
	}
	pw.rev.post(WireRevision{Steps: steps, Seg: seg})
}

func (pw *PacedWire) applyRevision(tick int64) {
	if pw == nil {
		return
	}
	rev, ok := pw.rev.take()
	if !ok {
		return
	}
	pw.ReviseInFlightGeometry(tick, rev.Steps, rev.Seg)
}

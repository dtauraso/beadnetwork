package bead

import "github.com/dtauraso/wirefold/nodes/spatial"

type Revision struct {
	Steps int
	Seg   spatial.Segment
}

type revisionSlot struct {
	ch chan Revision
}

func newRevisionSlot() revisionSlot {
	return revisionSlot{ch: make(chan Revision, 1)}
}

func (s *revisionSlot) post(rev Revision) {
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

func (s *revisionSlot) take() (Revision, bool) {
	if s.ch == nil {
		return Revision{}, false
	}
	select {
	case rev := <-s.ch:
		return rev, true
	default:
		return Revision{}, false
	}
}

func (pw *BeadRun) PostGeom(steps int, seg spatial.Segment) {
	if pw == nil {
		return
	}
	pw.rev.post(Revision{Steps: steps, Seg: seg})
}

func (pw *BeadRun) applyRevision() {
	if pw == nil {
		return
	}
	rev, ok := pw.rev.take()
	if !ok {
		return
	}
	pw.ReviseGeometry(rev.Steps, rev.Seg)
}

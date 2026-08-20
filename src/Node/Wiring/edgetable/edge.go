package edgetable

import (
	"github.com/dtauraso/wirefold/src/Node/wire"
	"github.com/dtauraso/wirefold/src/Node/wire/outport"
)

type Edge struct {
	label string
	srcID string
	dstID string
	srcH  string
	dstH  string

	out  *outport.Out
	dest *wire.BeadRun
}

func New(label, srcID, dstID, srcHandle, dstHandle string) *Edge {
	return &Edge{label: label, srcID: srcID, dstID: dstID, srcH: srcHandle, dstH: dstHandle}
}

func (e *Edge) Label() string     { return e.label }
func (e *Edge) SrcID() string     { return e.srcID }
func (e *Edge) DstID() string     { return e.dstID }
func (e *Edge) SrcHandle() string { return e.srcH }
func (e *Edge) DstHandle() string { return e.dstH }

func (e *Edge) SetOut(out *outport.Out) { e.out = out }
func (e *Edge) Out() *outport.Out       { return e.out }

func (e *Edge) SetDest(dest *wire.BeadRun) { e.dest = dest }
func (e *Edge) Dest() *wire.BeadRun        { return e.dest }

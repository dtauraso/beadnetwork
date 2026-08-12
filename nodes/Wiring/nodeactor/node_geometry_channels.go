// node_geometry_channels.go — the channel-touching half of NodeGeometry's POST-CONSTRUCTION
// exported surface, split out of node_geometry_accessors.go
// (docs/planning/movedispatch-decomposition.md §20).
//
// The channel-touching members (msg.centerOut, msg.extIn, msg.neighborIn) stay unexported
// and are reached ONLY through the methods below that close over them — PollCenter,
// SendExternal, EnqueueSend, NeighborTrySend — mirroring §17's edgeMover rule exactly: no
// channel is ever exported, so no other package can ever send on this node's own inbox by
// naming the channel directly.
package nodeactor

import (
	"context"
	"fmt"

	"github.com/dtauraso/wirefold/nodes/Wiring/movemsg"
)

// NeighborTrySend returns the non-blocking try-send func FROM peer fromID TO this node —
// wrapping this node's own neighborIn[fromID] channel exactly like edgemover's
// TrySendFromSrc/TrySendFromDst wrap that actor's own channels, so the channel itself never
// crosses the package boundary. Called by package Wiring's move_dispatch_construct.go
// resolveDest closure (bound at construction, invoked later — every invocation runs on
// fromID's own driving goroutine, via that node's own flushPending). ok is false when this
// node has no dedicated channel for fromID (not a direct neighbor).
func (m *NodeGeometry) NeighborTrySend(fromID string) (func(movemsg.Msg) bool, bool) {
	ch, ok := m.msg.neighborIn[fromID]
	if !ok {
		return nil, false
	}
	return func(msg movemsg.Msg) bool {
		select {
		case ch <- msg:
			return true
		default:
			return false
		}
	}, true
}

// PollCenter drains this node's own centerOut channel non-blockingly, returning the latest
// pushed center (ApplyCenter's own latest-wins push) if one is waiting. Called ONLY from
// package Wiring's dispatch/gesture goroutine (moverRegistry.drainCenterMirror) — the sole
// reader of every node's centerOut channel; this node's own goroutine is the sole writer
// (ApplyCenter).
func (m *NodeGeometry) PollCenter() (vec3, bool) {
	select {
	case c := <-m.msg.centerOut:
		return c, true
	default:
		return vec3{}, false
	}
}

// SendExternal sends msg on this node's own dedicated external-entry channel (extIn) — the
// bare EXTERNAL-caller path (package Wiring's RootMove/drag, gesture.go's dragStart send),
// not a mover-to-mover send (those go through EnqueueSend onto a node's own pending/
// flushPending retry queue, never through this method). Blocking with a ctx-cancel escape
// hatch: without the ctx.Done() arm, a send into a torn-down/full extIn on shutdown parks
// the calling goroutine forever (this node's own Run loop has already returned on the same
// ctx cancel, so nothing will ever drain it). ctx is nil only in tests that build a bare
// MoveDispatch without Start — a nil Context's Done() channel would panic, so guard it and
// fall back to a plain blocking send there (matches prior behavior; no shutdown path
// exists in that setting anyway).
func (m *NodeGeometry) SendExternal(ctx context.Context, msg movemsg.Msg) {
	if ctx == nil {
		m.msg.extIn <- msg
		return
	}
	select {
	case m.msg.extIn <- msg:
	case <-ctx.Done():
	}
}

// TryRecvExternal returns the next message on this node's own external-entry inbox
// (extIn), non-blocking — the receive-side counterpart to SendExternal, needed only when
// no NodeMover.Run/PairNodeSelf.Step goroutine is draining this node's own inbox (a bare
// test construction with no driving goroutine started). Production drains extIn from
// inside Run/Step's own select loop, never through this method.
func (m *NodeGeometry) TryRecvExternal() (movemsg.Msg, bool) {
	select {
	case msg := <-m.msg.extIn:
		return msg, true
	default:
		return movemsg.Msg{}, false
	}
}

// EnqueueSend is THIS node's own non-blocking send: it appends the message to this node's
// own pending retry queue and attempts an immediate flush — never blocking the calling
// handler goroutine. Bound once per node, at construction, as this node's own
// m.msg.sendMove (package Wiring's move_dispatch_construct.go: `ng.WireMessaging(...,
// md.mr.EnqueueFor(ng), ...)`, where enqueueFor now just returns this method value) so
// every send this node's own handle performs — including the ones
// broadcastToEdgesAndPartners makes on this node's own behalf via SendMove() — goes
// through this node's own retry queue, never a raw blocking channel write and never a
// second node's queue.
//
// This absorbs what used to be package Wiring's mover_registry.go enqueueFor closure body
// (a direct external touch of msg.pending) into the actor itself
// (docs/planning/movedispatch-decomposition.md §20) — the same class of runtime,
// own-goroutine write §19 already left as a bare field write for quantOffset/pending, now
// made unreachable from outside this package by construction rather than by convention.
// (The test-only message-trace tap this method used to fire — msg.tap/SetMsgTap/
// MoveDispatch.tapToInstall — was removed in §35, docs/planning/movedispatch-decomposition.md:
// no test anywhere in the repo called SetMsgTap, so it was dead observability, not a
// load-bearing seam.)
func (m *NodeGeometry) EnqueueSend(destID string, msg movemsg.Msg) {
	m.msg.pending = append(m.msg.pending, pendingSend{destID: destID, msg: msg})
	m.flushPending()
	if len(m.msg.pending) > maxPendingSends {
		// Named causes, checked against flushPending's actual behaviour (not
		// guessed): an item whose destID doesn't resolve is DROPPED, not
		// retained (flushPending's `!ok` branch), so an unresolvable
		// destination can never grow this queue — it is deliberately not
		// named below. What CAN: (1) a peer whose own goroutine has
		// stopped draining its inbox entirely (wedged or dead) — every
		// later item to that same destination piles up behind it,
		// unattempted, to preserve FIFO; (2) this node enqueueing to a
		// live-but-slower peer faster, cycle over cycle, than that peer's
		// own goroutine drains its inbox — flushPending retries only ONE
		// send per blocked destination per cycle, so a persistent
		// per-cycle surplus accumulates even without a dead peer.
		panic(fmt.Sprintf(
			"NodeGeometry(%s): pending exceeded %d retry-queued sends; either a "+
				"destination's own goroutine has stopped draining its inbox "+
				"(wedged or dead), or this node is enqueueing to a peer faster "+
				"than that peer drains, cycle over cycle",
			m.id, maxPendingSends))
	}
}

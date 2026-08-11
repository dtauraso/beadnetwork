// stream_claim.go — the node stream's own claim-or-reject wrapper, a package-local
// duplicate of package Wiring's stream_claim.go (Wiring.claimedStream is unexported and
// its constructor is unexported ON PURPOSE, so this package cannot reuse that type
// directly) — the SAME duplication precedent §17 set for nodes/Wiring/edgemover's own
// stream_claim.go (movedispatch-decomposition.md §17/§20), not a third independent
// mechanism: same shape, same reasoning. A second claim for the same key is a wiring-time
// failure reported to stderr, never a panic, and the second claimant gets a dead
// StreamHandle (Write is then a safe no-op).
//
// Reusing edgemover.ClaimRegistry/StreamHandle directly (rather than duplicating again)
// was considered and declined: it would make this package depend on edgemover for a
// mechanism that has nothing to do with edges, coupling two sibling actor packages for no
// reason beyond avoiding an 18-line copy. Package Wiring's own three claim registries
// (node/edge/view) already prove the real invariant — disjoint (kind, key) namespaces
// mean splitting the registry by kind changes nothing observable — so a fourth, disjoint
// registry here costs nothing either (see node_geometry_wire.go's WireStream doc comment
// for the node-specific proof).
package nodeactor

import (
	"fmt"
	"io"
	"os"
)

// ClaimRegistry tracks which node ids have already claimed a stream. Allocated once
// (NewClaimRegistry) before any claim is made and never touched after wiring finishes —
// plain map, no lock: this all runs on ONE goroutine before any NodeMover/PairNodeSelf
// goroutine exists.
type ClaimRegistry map[string]bool

// NewClaimRegistry returns a fresh, empty claim registry.
func NewClaimRegistry() ClaimRegistry { return ClaimRegistry{} }

// StreamHandle wraps an io.Writer that has been routed through Claim's claim-or-reject
// check. It deliberately exposes only Write and Ok, keeping it a narrow capability instead
// of a transparent alias a caller could stash and hand to a second goroutine.
type StreamHandle struct {
	w io.Writer
}

// Claim claims key in reg for w, or returns a dead StreamHandle (and reports the collision
// to stderr) if key was already claimed. reg may be nil (test construction that doesn't
// care about double-claim detection); w may be nil (no dedicated fd for this node — the
// existing, required no-dedicated-stream fallback, unrelated to claiming).
func Claim(reg ClaimRegistry, key string, w io.Writer) StreamHandle {
	if reg != nil {
		if reg[key] {
			fmt.Fprintf(os.Stderr,
				"stream-claim collision: node stream %q already claimed; a second wiring call "+
					"cannot also claim it — handing it to a second goroutine would reintroduce "+
					"the two-goroutines-one-fd desync docs/investigations/interior-stream-framing.md documents. "+
					"This claimant's stream stays unwired (writes nothing) instead.\n",
				key)
			return StreamHandle{}
		}
		reg[key] = true
	}
	return StreamHandle{w: w}
}

// Ok reports whether this StreamHandle actually wraps a live writer.
func (h StreamHandle) Ok() bool { return h.w != nil }

// Write delegates to the wrapped writer, or is a safe no-op when this StreamHandle is dead.
func (h StreamHandle) Write(p []byte) (int, error) {
	if h.w == nil {
		return len(p), nil
	}
	return h.w.Write(p)
}

// stream_claim.go — the edge stream's own claim-or-reject wrapper, package-local
// duplicate of nodes/Wiring's stream_claim.go (Wiring.claimedStream is unexported and its
// constructor is unexported ON PURPOSE, so this package cannot reuse that type directly —
// see this move's own report for why duplicating this small mechanism, rather than
// exporting Wiring's, is the correct call). Same shape, same reasoning: a second claim for
// the same key is a wiring-time failure reported to stderr, never a panic, and the second
// claimant gets a dead StreamHandle (Write is then a safe no-op).

package edgemover

import (
	"fmt"
	"io"
	"os"
)

// ClaimRegistry tracks which edge labels have already claimed a stream. Allocated once
// (NewClaimRegistry) before any claim is made and never touched after wiring finishes —
// plain map, no lock: this all runs on ONE goroutine before any edgeMover goroutine exists.
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
// care about double-claim detection); w may be nil (no dedicated fd for this edge — the
// existing, required no-dedicated-stream fallback, unrelated to claiming).
func Claim(reg ClaimRegistry, key string, w io.Writer) StreamHandle {
	if reg != nil {
		if reg[key] {
			fmt.Fprintf(os.Stderr,
				"stream-claim collision: edge stream %q already claimed; a second wiring call "+
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

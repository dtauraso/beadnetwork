// view_claim.go — the VIEW stream's OWN claim registry, split out of Wiring's
// stream_claim.go (docs/planning/gesture-actor.md's lift): the VIEW stream's claim key was
// always the empty string ("view:" + "") — a singleton, distinct from the node/edge
// registry's "node:<id>"/"edge:<label>" keys, which stay in Wiring's streamWiring — so a
// second, physically separate registry here cannot collide with the node/edge one and needs
// no shared map. See docs/planning/gesture-actor.md's "splitting streamWiring is CLEAN" note.
package viewstate

import (
	"fmt"
	"io"
	"os"
)

// viewClaimedStream wraps an io.Writer that has been routed through newViewClaimedStream's
// claim-or-reject check — same shape as Wiring's claimedStream (stream_claim.go), scoped to
// this package's one singleton stream. Exposes only Write and Ok.
type viewClaimedStream struct {
	w io.Writer
}

// newViewClaimedStream is the VIEW stream's one claim site (SetViewStream below). claimed
// tracks whether this UIState's view stream has already been claimed once — a plain bool,
// not a map, since there is exactly one key ("view stream", singleton). w may be nil (no
// WIREFOLD_STREAM_FDS "view" entry — the existing no-dedicated-stream fallback, unrelated to
// claiming).
func newViewClaimedStream(claimed *bool, w io.Writer) viewClaimedStream {
	if claimed != nil {
		if *claimed {
			fmt.Fprintf(os.Stderr,
				"stream-claim collision: view stream already claimed; a second wiring call "+
					"cannot also claim it — handing it to a second goroutine would reintroduce "+
					"the two-goroutines-one-fd desync docs/investigations/interior-stream-framing.md documents. "+
					"This claimant's stream stays unwired (writes nothing) instead.\n")
			return viewClaimedStream{}
		}
		*claimed = true
	}
	return viewClaimedStream{w: w}
}

// Ok reports whether this viewClaimedStream actually wraps a live writer.
func (c viewClaimedStream) Ok() bool { return c.w != nil }

// Write delegates to the wrapped writer, or is a safe no-op when dead.
func (c viewClaimedStream) Write(p []byte) (int, error) {
	if c.w == nil {
		return len(p), nil
	}
	return c.w.Write(p)
}

// stream_claim.go — the structural (not audited) form of "one dedicated inherited-stdio
// pipe per emitting goroutine" (memory/feedback_no_single_writer_bridge.md) for the THREE
// per-mover/singleton streams that used to be assigned as a bare io.Writer with no
// protection at all: nodeMover.streamOut, edgeMover.streamOut, streamWiring.viewOut. (The
// FOURTH — interior — already meets this bar a different way: Wiring.DrivenOut's
// unexported field + unexported constructor makes handing a DriveHeld goroutine the node's
// own shared interior stream a COMPILE ERROR; see driven_out.go's header comment. This file
// is the same idea applied to the three streams that had no such wrapper.)
//
// Before this file, `em.streamOut = os.NewFile(...)` (stream_wiring.go) and
// `md.sw.viewOut = out` (view_stream.go) were plain field assignments: nothing stopped a
// SECOND call from silently overwriting the first (or, worse, a second wiring path handing
// the SAME underlying fd to a different goroutine) — exactly the "someone traced the call
// sites and there's only one today" claim that was true, documented, and then false for the
// interior stream (docs/interior-stream-framing.md). An audit describes today; it does not
// prevent tomorrow's second wiring call.
//
// claimedStream is the fix: its only field is unexported, and the only way to build one
// that actually writes is newClaimedStream, called ONLY from the three claim sites this
// file's callers own (setEdgeStreams/setNodeStreams in stream_wiring.go, SetViewStream in
// view_stream.go). newClaimedStream consults a streamClaims registry — plain single-
// threaded bookkeeping during LoadTopology's wiring phase, before any mover/gesture
// goroutine exists, mirroring BuildArgs.driveSlotClaims' existing precedent (no lock
// needed: this all runs on ONE goroutine before any other goroutine is launched, and no
// lock/atomic belongs in this network regardless — tools/check-no-network-locks.sh). A
// second claim for the same (kind, key) is a WIRING-TIME failure reported to stderr (never
// a panic — main.go's existing stream-fd-mismatch posture: loud, not a crash-loop), and the
// second claimant receives a DEAD claimedStream (zero value: Write is then a safe no-op),
// exactly like BuildArgs.DriveOut's second-claimant fallback.
package Wiring

import (
	"fmt"
	"io"
	"os"
)

// streamClaims tracks which (kind, key) pairs have already been claimed. Kind is one of
// "node"/"edge"/"view" (this file's three callers); key is the claiming row's stable id
// ("" for the view stream, which is a singleton). Allocated once (newStreamClaims) before
// any claim is made and never touched after wiring finishes — plain map, no lock, same
// shape and same justification as BuildArgs.driveSlotClaims.
type streamClaims map[string]bool

func newStreamClaims() streamClaims { return streamClaims{} }

// claimedStream wraps an io.Writer that has been routed through newClaimedStream's
// claim-or-reject check — see this file's header comment for why this type exists. It
// deliberately exposes only Write and Ok (not an Unwrap() back to the bare io.Writer),
// keeping it a narrow capability instead of a transparent alias a caller could stash and
// hand to a second goroutine under a different name (same posture as Wiring.DrivenOut).
type claimedStream struct {
	w io.Writer
}

// newClaimedStream is unexported ON PURPOSE — see this file's header comment. Its only
// PRODUCTION callers are setEdgeStreams/setNodeStreams (stream_wiring.go) and
// SetViewStream (view_stream.go), all in this same package. reg may be nil (test
// construction that doesn't care about double-claim detection); w may be nil (no
// WIREFOLD_STREAM_FDS entry for this kind — the existing, required no-dedicated-stream
// fallback, unrelated to claiming).
func newClaimedStream(reg streamClaims, kind, key string, w io.Writer) claimedStream {
	if reg != nil {
		claimKey := kind + ":" + key
		if reg[claimKey] {
			fmt.Fprintf(os.Stderr,
				"stream-claim collision: %s stream %q already claimed; a second wiring call "+
					"cannot also claim it — handing it to a second goroutine would reintroduce "+
					"the two-goroutines-one-fd desync docs/interior-stream-framing.md documents. "+
					"This claimant's stream stays unwired (writes nothing) instead.\n",
				kind, key)
			return claimedStream{}
		}
		reg[claimKey] = true
	}
	return claimedStream{w: w}
}

// Ok reports whether this claimedStream actually wraps a live writer — false covers BOTH
// the ordinary "no dedicated fd for this kind" fallback (w started nil) and a rejected
// second claim (this file's header comment).
func (c claimedStream) Ok() bool { return c.w != nil }

// Write delegates to the wrapped writer, or is a safe no-op when this claimedStream is
// dead (Ok() == false) — matches the fallback behaviour every nil-io.Writer call site in
// this package already relies on.
func (c claimedStream) Write(p []byte) (int, error) {
	if c.w == nil {
		return len(p), nil
	}
	return c.w.Write(p)
}

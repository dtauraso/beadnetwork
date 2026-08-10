// stdin_reader.go — reads FRAMED BINARY records from stdin and dispatches them.
//
// The editor→Go bridge is a purely BINARY buffer (symmetric with the Go→TS content
// buffer on fd 3): each message is a binary RECORD written FRAMED as [len:u32-LE][record]
// to stdin. input_codec.go decodes a record into the stdinMsg below; the dispatch switch
// calls into the three Handlers (ApplyEdit / HandleRawInput / HandleSave, package
// nodes/Wiring's ApplyEdit / HandleRawInputMsg / HandleSaveMsg wrapped as closures by the
// caller so this package does not import nodes/Wiring) — only the wire decode moved from
// newline-JSON to framed binary.
//
// ONE JOB HERE: bytes off the pipe, whole records out. The message SHAPES are
// stdin_msg_types.go, the decode is input_codec.go, and the routing tables (applyEdit /
// applyUpdate / the per-attr tables) are nodes/Wiring's stdin_dispatch.go. This file owns
// framing, back-pressure, and shutdown — nothing else.
//
// The editor→Go bridge carries these top-level message kinds (all fully binary; no JSON
// on the wire — see input_codec.go). This list is the AUTHORITATIVE doc for the dispatch
// switch below and is checked against it by tools/bridge/check-message-kind-parity.sh: every type
// fenced by MSG_TYPES_START/END must be declared here and vice versa. Adding a case without
// adding a numbered entry (or the reverse) fails the guard.
//
// MSG_TYPES_DOC_START
//
//  1. "edit" — geometry-CRUD. The sole op is update, which sets an ATTRIBUTE on a
//     typed entity; the live entities are overlays, clock, and distanceGroup:
//       update overlays attr=toggle: flip one named overlay flag.
//       update clock attr=speed: set the playback multiplier.
//       update distanceGroup attr=length: adjust one "distance home button" group's
//       target pair length (×1.1 up / ÷1.1 down) — see ApplyDistanceGroupTarget.
//     A create/delete op pair (records 20/21) once added or removed an edge by
//     destination slot; both were removed end-to-end (no live TS sender, and create's
//     only trigger tore down a live wire's beads via PacedWire.Restore) — records 20/21
//     are now GAPS. Camera / node-move / port-anchor are NOT edits: the gesture FSM
//     produces them in-process from raw-input, so they never cross this seam as an edit op.
//
//  2. "save" — Go persists its OWN authoritative scene state (overlay visibility →
//     overlays.json; camera → camera.json). Bare command, no payload; the
//     editor holds no authoritative scene document.
//
//  3. "raw-input" — a raw pointer/wheel event + stateless raycast hit, handed to the
//     gesture FSM.
//
// A remounted webview that has nothing new to render (Go idle) is served from the
// EXT HOST's cached last stream frame instead of asking Go to manufacture one — see
// runCommand.ts's BuildAndRunRunner.lastSnapshot/getLastSnapshot. Go has no "resend"
// concept: it emits a frame only when something changes, and that stays true here.
//
// MSG_TYPES_DOC_END
//
// Go owns the clock and delivery; nothing on this seam triggers delivery or
// carries animation internals.
//
// One goroutine; cancellable via context. On EOF or context cancel, exits
// cleanly. Unknown message types and ops are silently ignored (forward-compat).

package stdinreader

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/dtauraso/wirefold/nodes/Wiring/inputcodec"
)

// Handlers is the three dispatch operations RunStdinReader needs, passed as function
// values so this package does not import nodes/Wiring. The caller (runtopology) builds
// them from its own *Wiring.MoveDispatch, closing over whatever else each op needs
// (Trace, speedSinks, slotReg) — RunStdinReader itself only ever hands each one the
// decoded message.
type Handlers struct {
	// ApplyEdit handles a decoded "edit" message (geometry-CRUD update op).
	ApplyEdit func(msg inputcodec.StdinMsg)
	// HandleRawInput handles a decoded "raw-input" message (pointer/wheel + raycast hit
	// → gesture FSM).
	HandleRawInput func(msg inputcodec.StdinMsg)
	// HandleSave handles the bare "save" command.
	HandleSave func()
}

// RunStdinReader reads FRAMED BINARY records from r, dispatching geometry-CRUD "edit"
// messages and the bare save command. RunStdinReader itself returns
// when ctx is done or r reaches EOF. Its background frame-reader goroutine (which
// blocks in io.ReadFull) unwinds on ctx-done too: if r implements io.Closer,
// RunStdinReader closes it on ctx-done, which unblocks the parked read (returns an
// error) so the reader goroutine exits via its close(recCh); return path. In production
// r is os.Stdin and this only runs as the process is already exiting, so the close is
// harmless. Call in a goroutine alongside the node run loop.
//
// h supplies the three dispatch operations (Handlers, above); any nil field is simply
// never called (a message of that kind is then silently ignored, same as an unknown type).
// maxFrameBytes bounds a single framed-binary record: the reader buffer size and the
// upper limit a decoded [len:u32] is allowed to request, so a corrupt/hostile length can't
// drive an unbounded allocation. Matches the 1 MB headroom of the pre-frame line buffer.
const maxFrameBytes = 1 << 20

func RunStdinReader(ctx context.Context, r io.Reader, h Handlers) {
	// Every persister now writes synchronously the moment its value changes (see
	// scene_persist.go's header comment for why the prior debounce/clean-shutdown-flush
	// machinery was removed), so there is nothing pending to flush on exit here anymore.
	// Framed-binary reader: each record is [len:u32-LE][record bytes]. A background
	// goroutine reads whole frames (io.ReadFull handles partial reads — a frame split
	// across TCP/pipe chunks is reassembled before the record is decoded) and hands the
	// record bytes to the dispatch loop over a channel. The channel keeps the dispatch
	// ctx-aware exactly as the old line reader did.
	br := bufio.NewReaderSize(r, maxFrameBytes)
	done := ctx.Done()
	recCh := make(chan []byte, 8)
	// Unblock the background frame-reader's io.ReadFull on ctx-cancel even when r stays
	// open (no EOF): if r is an io.Closer, close it once ctx is done so the parked read
	// returns an error and the reader goroutine can exit via its close(recCh); return path.
	// In production r is os.Stdin and this runs only as the process is already exiting, so
	// closing it is harmless; the goroutine simply outlives it until then.
	if c, ok := r.(io.Closer); ok {
		go func() {
			<-done
			c.Close()
		}()
	}
	go func() {
		var lenBuf [4]byte
		for {
			if _, err := io.ReadFull(br, lenBuf[:]); err != nil {
				if err != io.EOF && err != io.ErrUnexpectedEOF {
					fmt.Fprintf(os.Stderr, "stdin_reader: frame-length read error: %v\n", err)
				}
				close(recCh)
				return
			}
			n := binary.LittleEndian.Uint32(lenBuf[:])
			// Cap the frame size to the same 1 MB headroom the old line buffer had, so a
			// corrupt/hostile length can't drive an unbounded allocation and deafen the bridge.
			if n == 0 || n > maxFrameBytes {
				fmt.Fprintf(os.Stderr, "stdin_reader: bad frame length %d; stopping reader\n", n)
				close(recCh)
				return
			}
			rec := make([]byte, n)
			if _, err := io.ReadFull(br, rec); err != nil {
				if err != io.EOF && err != io.ErrUnexpectedEOF {
					fmt.Fprintf(os.Stderr, "stdin_reader: frame body read error: %v\n", err)
				}
				close(recCh)
				return
			}
			select {
			case recCh <- rec:
			case <-done:
				return
			}
		}
	}()
	for {
		select {
		case <-done:
			return
		case rec, ok := <-recCh:
			if !ok {
				return
			}
			// Row-identity resolution (a "raw-input" record's rawHit carries only numeric
			// rows; portFromHit/edgeFromHit/nodeFromHit in gesture.go resolve them) reads
			// MoveDispatch's portRowTable/edgeRowTable/nodeRowTable directly, inside
			// h.HandleRawInput's closure — those are a LOAD-TIME CONSTANT built once in
			// newMoveDispatch (move_dispatch_construct.go buildRowTables), not a
			// per-iteration drain: node/edge/port row order never changes after load (a
			// new node/edge only ever arrives via a full respawn), so there is nothing to
			// drain here anymore. Likewise heldCenters/centerOfNode read the dispatch
			// goroutine's own centerMirror, kept current by message from each mover —
			// there is no accumulated positions map to drain either.
			msg, decoded := inputcodec.DecodeInputRecord(rec)
			if !decoded {
				continue
			}
			// The authoritative per-type doc is the MSG_TYPES_DOC block in this file's
			// header. check-message-kind-parity.sh holds the two in parity — do not
			// document a type only here, and do not add a case without a doc entry.
			// MSG_TYPES_START
			switch msg.Type {
			case "edit":
				if h.ApplyEdit != nil {
					h.ApplyEdit(msg)
				}
			case "raw-input":
				if h.HandleRawInput != nil {
					h.HandleRawInput(msg)
				}
			case "save":
				if h.HandleSave != nil {
					h.HandleSave()
				}
			}
			// MSG_TYPES_END
		}
	}
}

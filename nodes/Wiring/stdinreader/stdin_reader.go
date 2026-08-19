// fenced by MSG_TYPES_START/END must be declared here and vice versa. Adding a case without
// adding a numbered entry (or the reverse) fails the guard.

// MSG_TYPES_DOC_START
//
//  1. "edit" — geometry-CRUD. The sole op is update, which sets an ATTRIBUTE on a
//       update overlays attr=toggle: flip one named overlay flag.
//       update clock attr=speed: set the playback multiplier.
//     A create/delete op pair (records 20/21) once added or removed an edge by
//     destination slot; both were removed end-to-end (no live TS sender, and create's
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

type Handlers struct {
	ApplyEdit func(msg inputcodec.StdinMsg)

	HandleRawInput func(msg inputcodec.StdinMsg)

	HandleSave func()
}

const maxFrameBytes = 1 << 20

func RunStdinReader(ctx context.Context, r io.Reader, h Handlers) {

	br := bufio.NewReaderSize(r, maxFrameBytes)
	done := ctx.Done()
	recCh := make(chan []byte, 8)

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

			msg, decoded := inputcodec.DecodeInputRecord(rec)
			if !decoded {
				continue
			}
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

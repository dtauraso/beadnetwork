// MSG_TYPES_DOC_START
//
//  1. "edit" — geometry-CRUD; the sole op is update, which sets an ATTRIBUTE on a
//     typed entity.
//
//  2. "save" — Go persists its OWN authoritative scene state. Bare command, no payload.
//
//  3. "raw-input" — a key or delete press, a one-shot command queued in order. A pointer
//     or wheel event is the current input and crosses as a file, never here.
//
// MSG_TYPES_DOC_END

package Drag

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Handlers struct {
	ApplyEdit func(op string, entity, attr byte, payload []byte)

	HandleSave func()

	HandleCommand func(ev RawInputMsg)
}

const (
	kindSave       = 4
	kindRawInput   = 10
	KindEditUpdate = 22
)

func recordKind(rec []byte) string {
	if len(rec) == 0 {
		return ""
	}
	switch rec[0] {
	case kindSave:
		return "save"
	case kindRawInput:
		return "raw-input"
	case KindEditUpdate:
		return "edit"
	}
	return ""
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

			// MSG_TYPES_START
			switch recordKind(rec) {
			case "edit":
				if h.ApplyEdit != nil && len(rec) >= 3 {
					h.ApplyEdit("update", rec[1], rec[2], rec[3:])
				}
			case "raw-input":
				ev, ok := DecodeRawInput(rec)
				if ok && IsCommandKind(ev.Kind) {
					if h.HandleCommand != nil {
						h.HandleCommand(ev)
					}
					break
				}
				fmt.Fprintf(os.Stderr,
					"stdin_reader: a %q raw-input record arrived on stdin, but only %v cross here; "+
						"pointer input is the current input and crosses as view/input/<kind>.bin, so the sender was not updated and this event is dropped rather than queued\n",
					ev.Kind, CommandKinds)
			case "save":
				if h.HandleSave != nil {
					h.HandleSave()
				}
			}
			// MSG_TYPES_END
		}
	}
}

var UpdateKinds = []string{
	"overlays",
	"clock",
	"scene",
	"tiltVector",
	"panels",
	"node",
	"edge",
}

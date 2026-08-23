// MSG_TYPES_DOC_START
//
//  1. "edit" — geometry-CRUD; the sole op is update, which sets an ATTRIBUTE on a
//     typed entity.
//
//  2. "save" — Go persists its OWN authoritative scene state. Bare command, no payload.
//
//  3. "raw-input" — a raw pointer/wheel event plus a stateless raycast hit, handed to
//     the gesture FSM.
//
// MSG_TYPES_DOC_END

package inputactor

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
}

const (
	kindSave       = 4
	kindRawInput   = 10
	kindEditUpdate = 22
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
	case kindEditUpdate:
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
				// Raw input is the current input, and it arrives as a file the
				// gesture goroutine reads when it wakes. A record here means
				// something is still sending it down the pipe, where it would
				// queue and replay after the fingers stop.
				fmt.Fprintf(os.Stderr,
					"stdin_reader: a raw-input record arrived on stdin, but raw input crosses as %s; "+
						"the sender was not updated and this event is dropped rather than queued\n",
					"view/input/current.bin")
			case "save":
				if h.HandleSave != nil {
					h.HandleSave()
				}
			}
			// MSG_TYPES_END
		}
	}
}

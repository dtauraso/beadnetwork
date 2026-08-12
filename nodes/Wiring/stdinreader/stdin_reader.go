// fenced by MSG_TYPES_START/END must be declared here and vice versa. Adding a case without

// MSG_TYPES_DOC_START

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
			// The authoritative per-type doc is the MSG_TYPES_DOC block in this file's

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

package streamclaim

import (
	"fmt"
	"io"
	"os"
)

type ClaimRegistry map[string]bool

func NewClaimRegistry() ClaimRegistry { return ClaimRegistry{} }

type StreamHandle struct {
	w io.Writer
}

func Claim(reg ClaimRegistry, key string, w io.Writer) StreamHandle {
	if reg != nil {
		if reg[key] {
			fmt.Fprintf(os.Stderr,
				"stream-claim collision: node stream %q already claimed; a second wiring call "+
					"cannot also claim it — handing it to a second goroutine would reintroduce "+
					"the two-goroutines-one-fd desync that interleaves two frames' header and payload "+
					"writes into garbage (CLAUDE.md: one dedicated pipe per emitting goroutine). "+
					"This claimant's stream stays unwired (writes nothing) instead.\n",
				key)
			return StreamHandle{}
		}
		reg[key] = true
	}
	return StreamHandle{w: w}
}

func (h StreamHandle) Ok() bool { return h.w != nil }

func (h StreamHandle) Write(p []byte) (int, error) {
	if h.w == nil {
		return len(p), nil
	}
	return h.w.Write(p)
}

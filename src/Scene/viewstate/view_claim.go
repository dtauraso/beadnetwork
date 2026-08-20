package viewstate

import (
	"fmt"
	"io"
	"os"
)

type viewClaimedStream struct {
	w io.Writer
}

func newViewClaimedStream(claimed *bool, w io.Writer) viewClaimedStream {
	if claimed != nil {
		if *claimed {
			fmt.Fprintf(os.Stderr,
				"stream-claim collision: view stream already claimed; a second wiring call "+
					"cannot also claim it — handing it to a second goroutine would reintroduce "+
					"the two-goroutines-one-fd desync that interleaves two frames' header and payload "+
					"writes into garbage (CLAUDE.md: one dedicated pipe per emitting goroutine). "+
					"This claimant's stream stays unwired (writes nothing) instead.\n")
			return viewClaimedStream{}
		}
		*claimed = true
	}
	return viewClaimedStream{w: w}
}

func (c viewClaimedStream) Ok() bool { return c.w != nil }

func (c viewClaimedStream) Write(p []byte) (int, error) {
	if c.w == nil {
		return len(p), nil
	}
	return c.w.Write(p)
}

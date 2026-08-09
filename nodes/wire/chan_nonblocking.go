// chan_nonblocking.go — generic non-blocking channel operations.
//
// These are plain channel utilities (latest-wins send, drain-the-latest
// receive). They are not port behaviour: they know nothing about In / Out /
// PacedWire and would read identically for any channel of the same element
// type. They live in their own file so ports.go stays about the ports.

package wire

// drainStepsNonBlocking folds the latest pending value off ch (if any) into
// *cur, without blocking. A nil ch (chan-mode Out, or an unpublished port)
// simply never selects the receive case, leaving *cur at its zero value.
func drainStepsNonBlocking(ch chan int, cur *int) {
	select {
	case v := <-ch:
		*cur = v
	default:
	}
}

// sendIntNonBlocking delivers v to ch, latest-wins: if the buffer already holds
// an undrained stale value, that stale value is dropped and replaced — mirrors
// SendSpeedNonBlocking (clock.go) for the same reason (absolute state, not an
// event stream). A nil ch (chan-mode Out) makes every case here select
// `default`, so this is a silent no-op.
func sendIntNonBlocking(ch chan int, v int) {
	select {
	case ch <- v:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- v:
	default:
	}
}

// drainSegNonBlocking is drainStepsNonBlocking's WireSegment counterpart.
func drainSegNonBlocking(ch chan WireSegment, start, end *Vec3) {
	select {
	case seg := <-ch:
		*start, *end = seg.Start, seg.End
	default:
	}
}

// sendSegNonBlocking is sendIntNonBlocking's WireSegment counterpart.
func sendSegNonBlocking(ch chan WireSegment, seg WireSegment) {
	select {
	case ch <- seg:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- seg:
	default:
	}
}

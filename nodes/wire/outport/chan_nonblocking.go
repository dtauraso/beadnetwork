package outport

import "github.com/dtauraso/wirefold/nodes/wire"

func drainStepsNonBlocking(ch chan int, cur *int) {
	select {
	case v := <-ch:
		*cur = v
	default:
	}
}

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

func drainSegNonBlocking(ch chan wire.WireSegment, start, end *wire.Vec3) {
	select {
	case seg := <-ch:
		*start, *end = seg.Start, seg.End
	default:
	}
}

func sendSegNonBlocking(ch chan wire.WireSegment, seg wire.WireSegment) {
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

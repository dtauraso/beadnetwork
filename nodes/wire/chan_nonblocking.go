package wire

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

func drainSegNonBlocking(ch chan WireSegment, start, end *Vec3) {
	select {
	case seg := <-ch:
		*start, *end = seg.Start, seg.End
	default:
	}
}

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

package clock

func SendSleepMsNonBlocking(ch chan int64, ms int64) {
	select {
	case ch <- ms:
		return
	default:
	}
	select {
	case <-ch:
	default:
	}
	select {
	case ch <- ms:
	default:
	}
}

func RecvSleepMsNonBlocking(ch <-chan int64) (int64, bool) {
	select {
	case ms := <-ch:
		return ms, true
	default:
		return 0, false
	}
}

func ApplySpeedNonBlocking(clk Clock, speedCh <-chan float64) {
	select {
	case sp := <-speedCh:
		if rc, ok := clk.(*RealClock); ok {
			rc.SetSpeed(sp)
		}
	default:
	}
}

func RecvSpeedNonBlocking(speedCh <-chan float64) (float64, bool) {
	select {
	case sp := <-speedCh:
		return sp, true
	default:
		return 0, false
	}
}

func SendSpeedNonBlocking(ch chan float64, speed float64) {
	select {
	case ch <- speed:
		return
	default:
	}

	select {
	case <-ch:
	default:
	}
	select {
	case ch <- speed:
	default:
	}
}

func SendLatestNonBlocking(ch chan int64, v int64) {
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

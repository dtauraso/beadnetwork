package clock

import "time"

const SpeedNumScale = 4

func (c *RealClock) SetSpeed(speed float64) {
	if speed < 0 {
		speed = 0
	}
	now := time.Now()
	c.accScaled += time.Duration(float64(now.Sub(c.lastChange)) * c.speed)
	c.lastChange = now
	c.speed = speed
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
	sendLatest(ch, speed)
}

func SendSleepMsNonBlocking(ch chan int64, ms int64) {
	SendLatestNonBlocking(ch, ms)
}

func SendLatestNonBlocking(ch chan int64, v int64) {
	sendLatest(ch, v)
}

func sendLatest[T any](ch chan T, v T) {
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

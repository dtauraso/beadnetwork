package Speed

import (
	"github.com/dtauraso/wirefold/Categories/Clock"
)

const SpeedNumScale = 4

var attrSpeed = attrIndex("speed")

type Edit struct {
	Num int
}

func DecodeUpdate(payload []byte, attr byte) (Edit, bool) {
	r := NewReader(payload, 0)
	if attr != attrSpeed {
		return Edit{}, false
	}
	speed, err := r.U8()
	if err != nil {
		return Edit{}, false
	}
	return Edit{Num: int(speed)}, true
}

type Broadcaster interface {
	SendSpeed(num, divisor int64)
}

type SpeedState struct {
	Speed   *float64
	Divisor float64
}

func EditSpeed(e Edit, st SpeedState, sinks Broadcaster, persist func(float64), redraw func()) {
	SetSpeedNum(int64(e.Num), st, sinks, persist, redraw)
}

func SetSpeedNum(num int64, st SpeedState, sinks Broadcaster, persist func(float64), redraw func()) {
	divisor := int64(st.Divisor)
	if divisor < 1 {
		divisor = 1
	}
	sinks.SendSpeed(num, divisor)

	userSpeed := float64(num) / SpeedNumScale
	*st.Speed = userSpeed
	persist(userSpeed)
	redraw()
}

func ApplySpeedNonBlocking(clk clock.Clock, speedCh <-chan float64) {
	select {
	case sp := <-speedCh:
		if rc, ok := clk.(*clock.RealClock); ok {
			rc.SetSpeed(sp)
		}
	default:
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

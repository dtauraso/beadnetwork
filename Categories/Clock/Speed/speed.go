package Speed

import (
	"github.com/dtauraso/wirefold/Categories/Clock"
	"github.com/dtauraso/wirefold/Categories/Input/Stdin"
)

const SpeedNumScale = 4

var attrSpeed = attrIndex("speed")

func init() { Stdin.RegisterUpdateDecoder("clock", decodeUpdate) }

func decodeUpdate(payload []byte, attr byte) (Stdin.StdinMsg, bool) {
	r := NewReader(payload, 0)
	if attr != attrSpeed {
		return Stdin.StdinMsg{}, false
	}
	speed, err := r.U8()
	if err != nil {
		return Stdin.StdinMsg{}, false
	}
	return Stdin.StdinMsg{Type: "edit", Op: "update", Kind: "clock", Attr: "speed", Num: int(speed)}, true
}

type Broadcaster interface {
	SendSpeed(num, divisor int64)
}

type SpeedState struct {
	Speed   *float64
	Divisor float64
}

func EditSpeed(msg Stdin.StdinMsg, st SpeedState, sinks Broadcaster, persist func(float64), redraw func()) {
	if msg.Attr != "speed" {
		return
	}
	SetSpeedNum(int64(msg.Num), st, sinks, persist, redraw)
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

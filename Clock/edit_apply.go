package clock

import (
	"github.com/dtauraso/wirefold/Input/Stdin"
)

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

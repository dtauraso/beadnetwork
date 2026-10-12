package NodePhiTheta3

import "github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/CardPanel"

type Angles struct{ Phi, Theta int }

func ticksOf(v CardPanel.Vec, s int) Angles {
	return Angles{Phi: v.Phi * s, Theta: v.Theta * s}
}

func moveOne(start, end, turn, most, cw, ccw int) int {
	if cw^ccw != 1 {
		return start
	}
	gapCW := (end + turn - start) % turn
	gapCCW := (start + turn - end) % turn
	step := min(most, cw*gapCW+ccw*gapCCW)
	return (start + cw*step + ccw*(turn-step)) % turn
}

func moveToward(start, end Angles, s, m, cw, ccw int) Angles {
	turn := CardPanel.TurnTicks * s
	return Angles{
		Phi:   moveOne(start.Phi, end.Phi, turn, m*s, cw, ccw),
		Theta: moveOne(start.Theta, end.Theta, turn, m*s, cw, ccw),
	}
}

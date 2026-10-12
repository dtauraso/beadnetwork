package CardPanel

import (
	"fmt"
	"strconv"
)

const NodeCount = 3

const DefaultS = 30

const TurnTicks = 12

func SFits(s, turnPhi, turnTheta int) bool {
	ticks := TurnTicks * s
	return s >= 1 && turnPhi%ticks == 0 && turnTheta%ticks == 0
}

const DefaultM = 1

const (
	DefaultCW  = 1
	DefaultCCW = 0
)

type Vector int

const (
	VecStart Vector = iota
	VecEnd
	VecL
	VecS
	VecM
	VecCW
	VecCCW
)

func (v Vector) SceneWide() bool { return v == VecS || v == VecM || v == VecCW || v == VecCCW }

type Comp int

const (
	CompPhi Comp = iota
	CompTheta
	CompR
	CompOne
)

type Field struct {
	Node   int
	Vector Vector
	J      int
	Comp   Comp
}

type Vec struct{ Phi, Theta, R int }

func (v Vec) get(c Comp) int {
	switch c {
	case CompPhi:
		return v.Phi
	case CompTheta:
		return v.Theta
	case CompR:
		return v.R
	}
	panic(fmt.Sprintf("CardPanel.Vec.get: component %d is not phi, theta or r — a vector field was built with a scalar component", c))
}

func (v *Vec) set(c Comp, x int) {
	switch c {
	case CompPhi:
		v.Phi = x
	case CompTheta:
		v.Theta = x
	case CompR:
		v.R = x
	default:
		panic(fmt.Sprintf("CardPanel.Vec.set: component %d is not phi, theta or r — a vector field was built with a scalar component", c))
	}
}

type Card struct {
	Start Vec
	End   Vec
	L     [NodeCount]int
}

func (c Card) Get(f Field) int {
	switch f.Vector {
	case VecStart:
		return c.Start.get(f.Comp)
	case VecEnd:
		return c.End.get(f.Comp)
	case VecL:
		return c.L[f.J-1]
	}
	panic(fmt.Sprintf("CardPanel.Card.Get: vector %d is not a node vector — s, m, cw and ccw are scene-wide and are held beside the cards, not in one", f.Vector))
}

func (c *Card) Set(f Field, x int) {
	switch f.Vector {
	case VecStart:
		c.Start.set(f.Comp, x)
	case VecEnd:
		c.End.set(f.Comp, x)
	case VecL:
		c.L[f.J-1] = x
	default:
		panic(fmt.Sprintf("CardPanel.Card.Set: vector %d is not a node vector — s, m, cw and ccw are scene-wide and are held beside the cards, not in one", f.Vector))
	}
}

var compNames = map[Comp]string{CompPhi: "phi", CompTheta: "theta", CompR: "r"}

func (f Field) StateKey() string {
	switch f.Vector {
	case VecStart:
		return "start-" + compNames[f.Comp]
	case VecEnd:
		return "end-" + compNames[f.Comp]
	case VecL:
		return "l-" + strconv.Itoa(f.J)
	}
	panic(fmt.Sprintf("CardPanel.Field.StateKey: vector %d has no per-node state file — s, m, cw and ccw are persisted by the view owner", f.Vector))
}

func Partners(node int) [NodeCount - 1]int {
	var out [NodeCount - 1]int
	i := 0
	for j := 1; j <= NodeCount; j++ {
		if j != node {
			out[i] = j
			i++
		}
	}
	return out
}

func NodeFields(node int) []Field {
	out := []Field{
		{Node: node, Vector: VecStart, Comp: CompPhi},
		{Node: node, Vector: VecStart, Comp: CompTheta},
		{Node: node, Vector: VecStart, Comp: CompR},
		{Node: node, Vector: VecEnd, Comp: CompPhi},
		{Node: node, Vector: VecEnd, Comp: CompTheta},
	}
	for _, j := range Partners(node) {
		out = append(out, Field{Node: node, Vector: VecL, J: j, Comp: CompOne})
	}
	return out
}

var SharedFields = []Field{
	{Vector: VecS, Comp: CompOne}, {Vector: VecM, Comp: CompOne},
	{Vector: VecCW, Comp: CompOne}, {Vector: VecCCW, Comp: CompOne},
}

func CardFromState(node int, state map[string]int) Card {
	var c Card
	for _, f := range NodeFields(node) {
		c.Set(f, state[f.StateKey()])
	}
	return c
}

func Valid(f Field, x int) bool {
	switch f.Vector {
	case VecL, VecCW, VecCCW:
		return x == 0 || x == 1
	case VecS, VecM:
		return x >= 1
	case VecStart, VecEnd:
		if f.Comp == CompR {
			return x >= 0
		}
		return x >= 0 && x < TurnTicks
	}
	return true
}

type EditMsg struct {
	Field Field
	Value int
}

package CardPanel

import (
	"fmt"
	"strconv"
)

const NodeCount = 3

const DefaultS = 30

const TurnTicks = 12

const DefaultM = 1

type Vector int

const (
	VecStart Vector = iota
	VecPoleOffset
	VecK
	VecL
	VecS
	VecM
)

func (v Vector) SceneWide() bool { return v == VecS || v == VecM }

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
	Start      [NodeCount]Vec
	PoleOffset [NodeCount]Vec
	K          [NodeCount]int
	L          [NodeCount]int
}

func (c Card) Get(f Field) int {
	i := f.J - 1
	switch f.Vector {
	case VecStart:
		return c.Start[i].get(f.Comp)
	case VecPoleOffset:
		return c.PoleOffset[i].get(f.Comp)
	case VecK:
		return c.K[i]
	case VecL:
		return c.L[i]
	}
	panic(fmt.Sprintf("CardPanel.Card.Get: vector %d is not a node vector — s and m are scene-wide and are held beside the cards, not in one", f.Vector))
}

func (c *Card) Set(f Field, x int) {
	i := f.J - 1
	switch f.Vector {
	case VecStart:
		c.Start[i].set(f.Comp, x)
	case VecPoleOffset:
		c.PoleOffset[i].set(f.Comp, x)
	case VecK:
		c.K[i] = x
	case VecL:
		c.L[i] = x
	default:
		panic(fmt.Sprintf("CardPanel.Card.Set: vector %d is not a node vector — s and m are scene-wide and are held beside the cards, not in one", f.Vector))
	}
}

var compNames = map[Comp]string{CompPhi: "phi", CompTheta: "theta", CompR: "r"}

func (f Field) StateKey() string {
	j := strconv.Itoa(f.J)
	switch f.Vector {
	case VecStart:
		return "start-" + j + "-" + compNames[f.Comp]
	case VecPoleOffset:
		return "pole-offset-" + j + "-" + compNames[f.Comp]
	case VecK:
		return "k-" + j
	case VecL:
		return "l-" + j
	}
	panic(fmt.Sprintf("CardPanel.Field.StateKey: vector %d has no per-node state file — s and m are persisted by the view owner", f.Vector))
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
	var out []Field
	for _, j := range Partners(node) {
		out = append(out, Field{Node: node, Vector: VecStart, J: j, Comp: CompPhi})
		out = append(out, Field{Node: node, Vector: VecStart, J: j, Comp: CompTheta})
		out = append(out, Field{Node: node, Vector: VecStart, J: j, Comp: CompR})
	}
	for _, j := range Partners(node) {
		out = append(out, Field{Node: node, Vector: VecPoleOffset, J: j, Comp: CompPhi})
		out = append(out, Field{Node: node, Vector: VecPoleOffset, J: j, Comp: CompTheta})
	}
	for _, j := range Partners(node) {
		out = append(out, Field{Node: node, Vector: VecK, J: j, Comp: CompOne})
	}
	for j := 1; j <= NodeCount; j++ {
		out = append(out, Field{Node: node, Vector: VecL, J: j, Comp: CompOne})
	}
	return out
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
	case VecK, VecL:
		return x == 0 || x == 1
	case VecS, VecM:
		return x >= 1
	}
	return true
}

type EditMsg struct {
	Field Field
	Value int
}

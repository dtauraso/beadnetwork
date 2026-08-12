package distancegroups

import (
	"context"
	"time"

	"github.com/dtauraso/wirefold/nodes/spatial"
)

type Pair struct {
	Source string
	Target string
}

var GroupOrder = []string{"time", "input", "gate"}

var Groups = map[string][]Pair{
	"time":  {{Source: "2", Target: "5"}, {Source: "2", Target: "4"}, {Source: "4", Target: "7"}, {Source: "4", Target: "6"}},
	"input": {{Source: "1", Target: "3"}, {Source: "1", Target: "2"}},
	"gate":  {{Source: "3", Target: "8"}, {Source: "5", Target: "8"}, {Source: "5", Target: "9"}, {Source: "7", Target: "9"}},
}

func Max(hasGroups bool, centerOf func(string) (spatial.Vec3, bool), group string) (float64, bool) {
	if !hasGroups {
		return 0, false
	}
	pairs, ok := Groups[group]
	if !ok {
		return 0, false
	}
	max := 0.0
	any := false
	for _, p := range pairs {
		cs, okS := centerOf(p.Source)
		ct, okT := centerOf(p.Target)
		if !okS || !okT {
			continue
		}
		if d := ct.Sub(cs).Length(); d > max {
			max = d
		}
		any = true
	}
	return max, any
}

func Lens(hasGroups bool, centerOf func(string) (spatial.Vec3, bool)) (timeLen, inputLen, gateLen float32) {
	vals := make([]float32, len(GroupOrder))
	for i, g := range GroupOrder {
		if m, ok := Max(hasGroups, centerOf, g); ok {
			vals[i] = float32(m)
		}
	}
	return vals[0], vals[1], vals[2]
}

func ApplyTarget(ctx context.Context, hasGroups bool, centerOf func(string) (spatial.Vec3, bool), rootMove func(ctx context.Context, target string, newPos spatial.Vec3) bool, groupIdx, dir int) bool {
	if groupIdx < 0 || groupIdx >= len(GroupOrder) {
		return false
	}
	group := GroupOrder[groupIdx]
	pairs, ok := Groups[group]
	if !ok {
		return false
	}
	currentMax, any := Max(hasGroups, centerOf, group)
	if !any {
		return false
	}
	targetLen := currentMax * 1.1
	if dir < 0 {
		targetLen = currentMax / 1.1
	}
	moved := false
	for _, p := range pairs {
		cs, okS := centerOf(p.Source)
		ct, okT := centerOf(p.Target)
		if !okS || !okT {
			continue
		}
		offset := ct.Sub(cs)
		if offset.Length() == 0 {
			continue
		}
		newPos := cs.Add(offset.Normalize().Scale(targetLen))
		if rootMove(ctx, p.Target, newPos) {
			moved = true

			waitForCenterSettle(centerOf, p.Target, newPos)
		}
	}
	return moved
}

func waitForCenterSettle(centerOf func(string) (spatial.Vec3, bool), id string, want spatial.Vec3) {
	const tol = 1e-6
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if c, ok := centerOf(id); ok && c.Sub(want).Length() <= tol {
			return
		}
		time.Sleep(time.Millisecond)
	}
}

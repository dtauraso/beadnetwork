package Builder

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const DirStepClones = "clones"

func stepClonesDir(stepDir string, step int) string {
	return filepath.Join(stepDir, DirStepClones, "step-"+strconv.Itoa(step))
}

func (m Model) stepLinksFrom(step int, id string) []Links.Link {
	var out []Links.Link
	for _, l := range m.linksFrom(id) {
		if c := m.StepLinks[StepEdge{step, l.From, l.To}]; c != nil {
			l = *c
		}
		out = append(out, l)
	}
	return out
}

func (m *Model) SeedStepLinks() bool {
	if m.StepLinks == nil {
		m.StepLinks = map[StepEdge]*Links.Link{}
	}
	seeded := false
	for _, step := range StepNumbers {
		for _, l := range m.allLinks() {
			k := StepEdge{step, l.From, l.To}
			c := m.StepLinks[k]
			if c == nil {
				c := l
				m.StepLinks[k] = &c
				seeded = true
				continue
			}
			if c.Delta.S != l.Delta.S || c.Start.S != l.Start.S || c.Increment.S != l.Increment.S {
				c.SetS(l.Start.S)
				seeded = true
			}
		}
	}
	return seeded
}

func (m *Model) DropStepEdge(from, to string) error {
	for _, step := range StepNumbers {
		k := StepEdge{step, from, to}
		if m.StepPicks[k] {
			m.StepPicks[k] = false
		}
		delete(m.StepLinks, k)
		if m.StepDir != "" {
			if err := Links.RemoveLinksTo(filepath.Join(stepClonesDir(m.StepDir, step), to), from, to); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m Model) StepDeltas(step int) map[StepEdge]Links.Ints {
	out := map[StepEdge]Links.Ints{}
	for e, l := range m.StepLinks {
		if e.Step == step {
			out[e] = l.Delta
		}
	}
	return out
}

func (m Model) TurnFromDrawn(e StepEdge, was Links.Ints) (Links.Start, bool) {
	live, ok := m.Link(e.From, e.To)
	if !ok {
		return Links.Start{}, false
	}
	drawn := live.Dir()
	return Links.Start{Phi: Links.Wrap(1, drawn.Phi-float64(was.Phi)), Theta: Links.Wrap(1, drawn.Theta-float64(was.Theta)), R: drawn.R - float64(was.R), S: live.Start.S, One: live.Start.One}.Pulled(), true
}

func (m *Model) StepEdit(a Act, sign int, typed *float64) string {
	l := m.StepLinks[StepEdge{a.I, a.Node, a.To}]
	if l == nil {
		return fmt.Sprintf("step %d %s→%s: this step has no copy of the link yet, so nothing changed", a.I, a.Node, a.To)
	}
	name := fmt.Sprintf("step %d %s→%s", a.I, a.Node, a.To)
	switch a.Op {
	case OpLinkStart:
		v := map[string]float64{Links.KPhi: l.Start.Phi, Links.KTheta: l.Start.Theta, Links.KR: l.Start.R}[a.K] + float64(sign)
		if typed != nil && a.K == Links.KR {
			v = *typed
		} else if typed != nil {
			v = l.Start.FromTick(int(math.Floor(*typed)))
		}
		if a.K == Links.KR && (v < 0 || v+float64(l.Delta.R) < 0) {
			return fmt.Sprintf("%s r %v: neither r nor r + Δr can be negative, so r stays at %v", name, v, l.Start.R)
		}
		l.SetStart(a.K, v)
	case OpLinkOne:
		if sign == 0 {
			return ""
		}
		if a.K == Links.KR && l.Dir().R+float64(l.OneCarry(a.K, sign)) < 0 {
			return fmt.Sprintf("%s 1r: r + Δr would be below 0, so 1r stays at %d", name, l.Start.One.R)
		}
		l.StepOne(a.K, sign)
	case OpLinkPress:
		if sign == 0 {
			sign = a.Sign
		}
		if a.K == Links.KR && l.Start.R+float64(l.Delta.R+sign*l.Increment.R) < 0 {
			return fmt.Sprintf("%s Δr: r + Δr would be below 0, so Δr stays at %d", name, l.Delta.R)
		}
		l.Step(a.K, sign)
	case OpIncrement:
		inc := *l.Increment.At(a.K)
		v := inc + sign
		if typed != nil {
			if *typed != math.Trunc(*typed) {
				return fmt.Sprintf("%s nΔ %v: not a whole number, so it stays at %d", name, *typed, inc)
			}
			v = int(*typed)
		}
		if a.K == Links.KR && v <= 0 {
			return fmt.Sprintf("%s nΔr %d: it must stay above 0, so it stays at %d", name, v, inc)
		}
		*l.Increment.At(a.K) = v
	}
	return ""
}

func (m Model) SNeeded(id string) int {
	need := 1
	for _, l := range m.allLinks() {
		if l.From != id && l.To != id {
			continue
		}
		need = Links.Lcm(need, l.Start.TicksNeeded())
		for _, step := range StepNumbers {
			if c := m.StepLinks[StepEdge{step, l.From, l.To}]; c != nil {
				need = Links.Lcm(need, c.Start.TicksNeeded())
			}
		}
	}
	return need
}

func IsStepEdit(a Act) bool {
	return a.I > 0 && (a.Op == OpLinkStart || a.Op == OpLinkOne || a.Op == OpLinkPress || a.Op == OpIncrement)
}

func (m *Model) readStepLinks() {
	m.StepLinks = map[StepEdge]*Links.Link{}
	for _, step := range StepNumbers {
		dir := stepClonesDir(m.StepDir, step)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() { // path-resolution-ok: each step clone holds one directory per target node
				continue
			}
			ls, err := Links.ReadLinks(filepath.Join(dir, e.Name()), e.Name())
			if err != nil {
				continue
			}
			for _, l := range ls {
				c := l
				m.StepLinks[StepEdge{step, l.From, l.To}] = &c
			}
		}
	}
}

func (m Model) writeStepLinks() error {
	byTo := map[int]map[string][]Links.Link{}
	for k, l := range m.StepLinks {
		if byTo[k.Step] == nil {
			byTo[k.Step] = map[string][]Links.Link{}
		}
		byTo[k.Step][k.To] = append(byTo[k.Step][k.To], *l)
	}
	for step, tos := range byTo {
		for to, ls := range tos {
			if err := Links.WriteBaseLinks(filepath.Join(stepClonesDir(m.StepDir, step), to), ls); err != nil {
				return err
			}
		}
	}
	return nil
}

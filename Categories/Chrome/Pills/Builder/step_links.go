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
			if m.StepLinks[k] == nil {
				c := l
				m.StepLinks[k] = &c
				seeded = true
			}
		}
	}
	return seeded
}

func (m Model) StepDeltas(step int) map[StepEdge]Links.Ticks {
	out := map[StepEdge]Links.Ticks{}
	for e, l := range m.StepLinks {
		if e.Step == step {
			out[e] = l.Delta
		}
	}
	return out
}

func (m Model) TurnFromDrawn(e StepEdge, was Links.Ticks) (Links.Start, bool) {
	live, ok := m.Link(e.From, e.To)
	if !ok {
		return Links.Start{}, false
	}
	s := m.ScaleOf(e.To)
	drawn := live.Dir(s)
	return Links.Start{Phi: Links.Wrap(s, drawn.Phi-float64(was.Phi)), Theta: Links.Wrap(s, drawn.Theta-float64(was.Theta)), R: drawn.R - float64(was.R)}, true
}

func (m *Model) StepEdit(a Act, sign int, typed *float64) string {
	l := m.StepLinks[StepEdge{a.I, a.Node, a.To}]
	if l == nil {
		return fmt.Sprintf("step %d %s→%s: this step has no copy of the link yet, so nothing changed", a.I, a.Node, a.To)
	}
	s := m.ScaleOf(a.To)
	name := fmt.Sprintf("step %d %s→%s", a.I, a.Node, a.To)
	switch a.Op {
	case OpLinkStart:
		v := map[string]float64{Links.KPhi: l.Start.Phi, Links.KTheta: l.Start.Theta, Links.KR: l.Start.R}[a.K] + float64(sign)
		if typed != nil {
			v = *typed
		}
		if a.K == Links.KR && (v < 0 || v+float64(l.Delta.R) < 0) {
			return fmt.Sprintf("%s r %v: neither r nor r + Δr can be negative, so r stays at %v", name, v, l.Start.R)
		}
		l.SetStart(s, a.K, v)
	case OpLinkPress:
		if sign == 0 {
			sign = a.Sign
		}
		if a.K == Links.KR && l.Start.R+float64(l.Delta.R+sign*l.Increment.R) < 0 {
			return fmt.Sprintf("%s Δr: r + Δr would be below 0, so Δr stays at %d", name, l.Delta.R)
		}
		l.Step(s, a.K, sign)
	case OpIncrement:
		inc := map[string]int{Links.KPhi: l.Increment.Phi, Links.KTheta: l.Increment.Theta, Links.KR: l.Increment.R}[a.K]
		v := float64(inc + sign)
		if typed != nil {
			v = *typed
		}
		if v != math.Trunc(v) || (a.K == Links.KR && v <= 0) {
			return fmt.Sprintf("%s increment %v: not a whole number of ticks, so it stays at %d", name, v, inc)
		}
		switch a.K {
		case Links.KTheta:
			l.Increment.Theta = int(v)
		case Links.KR:
			l.Increment.R = int(v)
		default:
			l.Increment.Phi = int(v)
		}
	}
	return ""
}

func IsStepEdit(a Act) bool {
	return a.I > 0 && (a.Op == OpLinkStart || a.Op == OpLinkPress || a.Op == OpIncrement)
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

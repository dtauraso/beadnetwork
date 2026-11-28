package Builder

import (
	"path/filepath"
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard/Fill"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/StepStore"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
)

type StepNode struct {
	Step int
	ID   string
}

func (m Model) activeAt(p []int) []StepEdge {
	var out []StepEdge
	for _, l := range m.allLinks() {
		if m.amountAt(p, Edge{From: l.From, To: l.To}).Picked {
			out = append(out, StepEdge{p[len(p)-1], l.From, l.To})
		}
	}
	return out
}

func (m Model) stepActive(step int) []StepEdge {
	if m.StepIsParent(step) {
		return nil
	}
	return m.activeAt(m.stepPath(step))
}

func IsStepHead(a Act) bool {
	return a.I > 0 && (a.Op == OpAllLocks || a.Op == OpAllBoth || a.Op == OpAllLock || a.Op == OpAllLockCycle || a.Op == OpAllInc || a.Op == OpAllOne || a.Op == OpAllScale || a.Op == OpAllPlay)
}

func (m *Model) StepHeadEdit(a Act, sign int) (string, []Links.Envelope, error) {
	st := m.StepHead(a.I)
	switch a.Op {
	case OpAllLockCycle:
		way, on := st.At(a.K)
		switch {
		case !*on:
			*on, *way = true, 1
		case *way > 0:
			*way = -1
		default:
			*on, *way = false, 1
		}
		st.Both = st.Both && !*on
		return "", nil, nil
	case OpAllLocks:
		st.ToggleBoth()
		return "", nil, nil
	case OpAllScale:
		st.Scale = max(1, st.PlayScale()+sign)
		return "", nil, nil
	case OpAllInc, OpAllOne:
		return m.headEdit(a.I, a.Op, a.K, sign, Links.Lock{})
	case OpAllLock:
		way, on := st.At(a.K)
		switch {
		case sign == 0:
			*on = !*on
			st.Both = st.Both && !*on
		case *on:
			*way = sign
			return m.headEdit(a.I, OpAllLock, a.K, sign, st.Locks())
		default:
			return m.headEdit(a.I, OpAllLock, a.K, sign, st.Locks())
		}
		if *way == 0 {
			*way = 1
		}
		return "", nil, nil
	}
	if sign == 0 {
		sign = a.Sign
	}
	return m.headEdit(a.I, OpAllBoth, Links.KBoth, sign, st.Locks())
}

const (
	OpAllPlay  = "all-play"
	OpAllScale = "all-scale"
)

type Play struct {
	Path []int
	K, N int
	At   int64
}

func (m Model) presses(p []int) int {
	n := 0
	for _, e := range m.activeAt(p) {
		n = max(n, m.amountAt(p, Edge{From: e.From, To: e.To}).Presses())
	}
	return n
}

func (m Model) PlayPresses(step int) int { return m.presses(m.stepPath(step)) }

func (m Model) playParts(p []int, press, of int) []Links.Envelope {
	var out []Links.Envelope
	for _, e := range m.activeAt(p) {
		out = append(out, Links.Envelope{To: e.To, Msg: Links.PlayPart{From: e.From, Press: press, Of: of}})
	}
	return out
}

func (m Model) Playing(step int) bool {
	return m.Play != nil && slices.Equal(m.Play.Path, m.stepPath(step))
}

func playOf(m Model, step int) Fill.Lock {
	glyph := "▶"
	if m.Playing(step) {
		glyph = "■ " + itoa(m.Play.K) + "/" + itoa(m.Play.N)
	}
	return Fill.Lock{Label: glyph, Axis: "×" + itoa(m.HeadOf(step).PlayScale()), Value: itoa(m.HeadOf(step).PlayScale()), On: m.Playing(step),
		Cycle: Act{Op: OpAllPlay, I: step}, Press: Act{Op: OpAllScale, I: step}}
}

func (m Model) scaleAt(p []int) int {
	level := StepStore.LevelOf(p[:len(p)-1])
	if level == m.Level {
		return m.HeadOf(p[len(p)-1]).PlayScale()
	}
	return StepStore.ReadStep(m.stepDir(p)).PlayScale()
}

func (m *Model) startRun(p []int, now int64) []Links.Envelope {
	before := m.upTo(p, false)
	before = slices.DeleteFunc(before, func(l []int) bool { return slices.Equal(l, p) })
	out := m.sends(func(e Edge) (Links.Ints, bool) { return m.sumOver(before, e), m.amountAt(p, e).Picked })
	for _, e := range m.activeAt(p) {
		edge := Edge{From: e.From, To: e.To}
		d := m.amountAt(p, edge).D
		if d.Zero() {
			continue
		}
		out = append(out, Links.Envelope{To: e.To, Msg: Links.Amount{From: e.From, D: d, Play: true}})
		m.Sent[edge] = m.Sent[edge].Add(d)
	}
	if level := StepStore.LevelOf(p[:len(p)-1]); level == m.Level {
		m.ShownStep, m.StepClosed = p[len(p)-1], false
	}
	m.Play = &Play{Path: slices.Clone(p), N: max(1, m.presses(p)*m.scaleAt(p)), At: now + 2}
	return out
}

func (m *Model) StartPlay(step int, now int64) []Links.Envelope {
	if m.StepIsParent(step) {
		return m.runFolder(m.leavesUnder(m.stepPath(step)), now)
	}
	return m.startRun(m.stepPath(step), now)
}

const FileNetworkOpen = "network-open.bin"

func (m *Model) ReadNetworkOpen() {
	m.NetworkOpen = false
	Store.ReadIfExists(filepath.Join(StepStore.PressDir(m.Root), FileNetworkOpen), &m.NetworkOpen)
}

func (m *Model) SetNetworkOpen(on bool) error {
	m.NetworkOpen = on
	return Store.WriteAtomicIfChanged(filepath.Join(StepStore.PressDir(m.Root), FileNetworkOpen), on)
}

func (m *Model) ReadNodeSteps() {
	for _, id := range m.Nodes {
		m.NodeStep(id)
	}
	if m.Root != "" {
		m.ReadNumbers()
		m.readStepsOpen()
		for _, step := range m.Numbers() {
			m.StepHead(step)
		}
	}
	m.ReadAmounts()
}

package Builder

import (
	"slices"
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

func stepBar(m Model) []Group {
	n := m.StepShown()
	if n == 0 {
		return nil
	}
	name := Field("step "+itoa(n), m.StepNames[n], Act{Op: OpStepName, I: n}, true, false).Plain().Wording().Tipped("name this step")
	bar := Group{Rows: [][]Cell{{
		name,
		Button("‹", Act{Op: OpStepEarlier, I: n}, false).Tipped("move this step one earlier"),
		Button("›", Act{Op: OpStepLater, I: n}, false).Tipped("move this step one later"),
		Button("→ in", Act{Op: OpStepEnter, I: n}, false).Tipped("go into this step's own steps"),
		Button("+", Act{Op: OpStepAdd, I: n}, false).Tipped("add a step after this one"),
		Button("⧉", Act{Op: OpStepCopy, I: n}, false).Tipped("duplicate this step"),
		Button("×", Act{Op: OpStepDrop, I: n}, false).Tipped("delete this step"),
	}}}
	return []Group{bar, headRow(m, n, 0)}
}

func signed(n int) string {
	if n > 0 {
		return "+" + strconv.Itoa(n)
	}
	return strings.Replace(strconv.Itoa(n), "-", "−", 1)
}

func turnText(l Links.Link) string {
	var parts []string
	for _, a := range axes {
		if d := a.delta(l); d != 0 {
			parts = append(parts, signed(d)+a.label)
		}
	}
	if len(parts) == 0 {
		return "0"
	}
	return strings.Join(parts, " ")
}

func startText(o Links.Start) string {
	return "φ" + strconv.Itoa(o.Tick.Phi) + " θ" + strconv.Itoa(o.Tick.Theta) + " r" + tenths(o.R)
}

func locksText(m Model, step int) string {
	st := m.headOf(step)
	var parts []string
	for _, a := range axes {
		way, on := st.At(a.k)
		glyph := "·"
		if *on && *way > 0 {
			glyph = "↑"
		} else if *on && *way < 0 {
			glyph = "↓"
		}
		parts = append(parts, a.label+glyph)
	}
	return strings.Join(parts, " ")
}

func scoreCell(m Model, step int, l Links.Link) Cell {
	e := StepEdge{step, l.From, l.To}
	sl, ok := m.gridLink(e)
	if !ok {
		return Text("").Dimmed(true)
	}
	if !sl.Picked {
		return LitText("— "+pickMark(false), cellAct(step, l), m.ScoreCell == e).Tipped("select to edit")
	}
	return LitText(turnText(sl)+" "+pickMark(true), cellAct(step, l), m.ScoreCell == e).Tipped(pressSummary(sl)).Dialed(dialOf(sl))
}

func dialOf(l Links.Link) Sheet.Dial {
	if l.Delta.Phi == 0 && l.Delta.Theta != 0 {
		return axisDial(l, Links.KTheta)
	}
	return axisDial(l, Links.KPhi)
}

func axisDial(l Links.Link, k string) Sheet.Dial {
	if k == Links.KTheta {
		return Sheet.Dial{Axis: Sheet.DialTheta, Start: float32(l.Start.Theta) + float32(l.Total.Theta), Turn: float32(l.Delta.Theta)}
	}
	return Sheet.Dial{Axis: Sheet.DialPhi, Start: float32(l.Start.Phi) + float32(l.Total.Phi), Turn: float32(l.Delta.Phi)}
}

func scoreGroup(m Model) Group {
	ns := m.stepNums()
	head := []Cell{Text("link").Inked(inkHead), Text("START").Inked(inkHead)}
	locks := []Cell{Text(""), Text("")}
	for _, n := range ns {
		head = append(head, LitText(m.StepLabel(n), Act{Op: OpStepShow, I: n}, n == m.StepShown()).Tipped("show this step in the scene · hover to ghost its places"))
		locks = append(locks, Text(locksText(m, n)).Dimmed(true).Tipped("HEAD locks for this step"))
	}
	rows := [][]Cell{head, locks}
	for _, id := range m.Nodes {
		links := m.linksTo(id)
		if len(links) == 0 {
			continue
		}
		rows = append(rows, []Cell{Text("into " + id).Inked(inkHead)})
		for _, l := range links {
			pick := m.ScoreCell.Step
			if pick == 0 {
				pick = m.StepShown()
			}
			row := []Cell{LitText(l.From+" → "+l.To, cellAct(pick, l), false).Tipped("select this link · hover to mark its edge"), Text(startText(m.Grids[id].Base[l.From].Origin)).Dimmed(true)}
			for _, n := range ns {
				row = append(row, scoreCell(m, n, l))
			}
			rows = append(rows, row)
		}
	}
	return Group{Title: "score — links down, steps across · ● moves its target in that step", Rows: rows, Slides: true, Table: "score", Keys: m.ScoreFocused()}
}

const scoreKeysHint = "←→↑↓ move · Space pick · + − press Δ · 1 2 3 axis φ θ r · Esc leave"

func (m Model) ScoreFocused() bool { return !m.AtSeqs && m.ScoreCell.Step != 0 }

func (m Model) scoreAxis() string {
	if m.ScoreAxis == "" {
		return Links.KPhi
	}
	return m.ScoreAxis
}

func (m Model) scoreLinks() []Links.Link {
	var out []Links.Link
	for _, id := range m.Nodes {
		out = append(out, m.linksTo(id)...)
	}
	return out
}

func (m *Model) scoreKey(key string) Act {
	e := m.ScoreCell
	ls, ns := m.scoreLinks(), m.stepNums()
	li, si := slices.IndexFunc(ls, func(l Links.Link) bool { return l.From == e.From && l.To == e.To }), slices.Index(ns, e.Step)
	move := func(dl, ds int) {
		if li >= 0 && si >= 0 {
			l := ls[min(max(li+dl, 0), len(ls)-1)]
			m.ScoreCell = StepEdge{ns[min(max(si+ds, 0), len(ns)-1)], l.From, l.To}
		}
	}
	at := Act{Node: e.From, To: e.To, I: e.Step}
	switch key {
	case "ArrowUp":
		move(-1, 0)
	case "ArrowDown":
		move(1, 0)
	case "ArrowLeft":
		move(0, -1)
	case "ArrowRight":
		move(0, 1)
	case "1", "2", "3":
		m.ScoreAxis = map[string]string{"1": Links.KPhi, "2": Links.KTheta, "3": Links.KR}[key]
	case "Escape":
		m.ScoreCell = StepEdge{}
	case " ":
		at.Op = OpStepPick
		return at
	case "+", "=", "-":
		l, ok := m.gridLink(e)
		if !ok || !l.Picked || m.scoreAxis() == Links.KR && m.FixedR(l) {
			return Act{}
		}
		at.Op, at.K, at.Sign = OpLinkPress, m.scoreAxis(), map[bool]int{true: -1, false: 1}[key == "-"]
		return at
	}
	return Act{}
}

package Builder

import (
	"strconv"
	"strings"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
)

const (
	OpLinksView = "links-view"
	viewTree    = "tree"
	viewScore   = "score"
)

func (m *Model) SetView(a Act) { m.Score = a.K == viewScore }

func viewGroup(m Model) Group {
	return Group{Rows: [][]Cell{{
		Text("view").Dimmed(true),
		Button(viewTree, Act{Op: OpLinksView, K: viewTree}, !m.Score).Tipped("one folding list"),
		Button(viewScore, Act{Op: OpLinksView, K: viewScore}, m.Score).Tipped("links down, steps across · read-only for now"),
	}}}
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
		head = append(head, Text(m.StepLabel(n)).Inked(inkStrong))
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
			row := []Cell{Text(l.From + " → " + l.To), Text(startText(m.Grids[id].Base[l.From].Origin)).Dimmed(true)}
			for _, n := range ns {
				row = append(row, scoreCell(m, n, l))
			}
			rows = append(rows, row)
		}
	}
	return Group{Title: "score — links down, steps across · ● moves its target in that step", Rows: rows, Slides: true, Table: "score"}
}

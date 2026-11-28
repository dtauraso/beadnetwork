package Builder

import (
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	"github.com/dtauraso/beadnetwork/Categories/Chrome/Pills/Builder/LinksBoard"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links"
	"github.com/dtauraso/beadnetwork/Categories/Node/Links/Store"
)

const LinksLabel = "Links"

const (
	OpLinkStart = "link-start"
	OpLinkPress = "link-press"
	OpLinkOne   = "link-one"
	OpStepShow  = "step-show"
	OpNetFold   = "network-fold"
	OpMenu      = "menu"
)

const (
	OpFrame  = "frame"
	OpTickMs = "tick-ms"

	FrameRho    = "rho"
	FrameTphi   = "tphi"
	FrameTtheta = "ttheta"
	FrameCr     = "cr"
	FrameS      = "s"
)

func frameAct(k string) Act { return Act{Op: OpFrame, K: k} }

const (
	menuStep = "step"
)

func (m *Model) PressMenu(a Act) {
	if a.Op != OpMenu {
		if a.Op != OpTreeFind && a.Op != OpFrame {
			m.Menu = Act{}
		}
		return
	}
	if m.Menu == a {
		m.Menu = Act{}
		return
	}
	m.Menu = a
}

func (m Model) allLinks() []Links.Link {
	var out []Links.Link
	for _, id := range m.Nodes {
		out = append(out, m.Node(id).Links()...)
	}
	Store.SortLinks(out)
	return out
}

func (m Model) linksTo(id string) []Links.Link {
	var out []Links.Link
	for _, l := range m.allLinks() {
		if l.To == id {
			out = append(out, l)
		}
	}
	return out
}

func LinksPill(m Model) Pill {
	d := boardData(m)
	switch {
	case m.NewOpen:
		d.New = newNodeData(m)
	case m.NetworkOpen:
		d.Net = netData(m)
	}
	return Pill{Label: LinksLabel, Bottom: true, Board: func(box Sheet.Rect, sx, sy float32) Sheet.Board { return LinksBoard.Lay(box, sx, sy, d) }}
}

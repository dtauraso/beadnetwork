package Board

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
)

const ValueRelFile = "view/chrome/board.bin"

var BoardValueNames = []string{
	"markKind", "markX", "markY", "markW", "markH", "markClip",
	"markFill", "markEdge", "markInk", "markRadius", "markStroke",
	"markText", "markLen", "markPx", "markBold", "markAlpha", "markDash", "markCorner",
	"dialAxis", "dialStart", "dialTurn", "dialGhost", "dialInc", "dialLabels",
	"clipX", "clipY", "clipW", "clipH", "editing", "caretMark", "caretAt",
}

func ValueRelPath() string { return ValueRelFile }

type State struct {
	w *BlobWriter
}

func (s *State) Arm(sceneRoot string) {
	s.w = NewBlobWriter(filepath.Join(sceneRoot, filepath.FromSlash(ValueRelFile)), BoardValueNames)
}

func weight(m Mark) uint8 {
	if m.Weight != 0 {
		return m.Weight
	}
	return flag(m.Bold)
}

func flag(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func (s *State) Write(l *Laid, e Sheet.Edit) {
	w := s.w
	if w == nil {
		return
	}
	w.Begin()
	w.U8("editing", flag(e.Active && l != nil))
	if l != nil {
		caret := uint32(0)
		for i, m := range l.Marks {
			if e.Active && m.Edit.Op != "" && m.Edit == e.Act {
				if m.Kind == KBox {
					m.Fill, m.Edge = On, Accent
				} else {
					m.Text, caret = e.Draft, uint32(i+1)
					if m.Align == Right {
						wt := weight(m)
						if wt == 1 {
							wt = 7
						}
						nw := TextW(m.Text, m.Px, wt)
						m.R.X, m.R.W = m.R.X+m.R.W-nw, nw
					}
					m.Ink = Ink
				}
			}
			w.U8("markKind", uint8(m.Kind))
			w.F32("markX", m.R.X)
			w.F32("markY", m.R.Y)
			w.F32("markW", m.R.W)
			w.F32("markH", m.R.H)
			w.U8("markClip", m.Clip)
			w.U8("markFill", uint8(m.Fill))
			w.U8("markEdge", uint8(m.Edge))
			w.U8("markInk", uint8(m.Ink))
			w.F32("markRadius", m.Radius)
			w.F32("markStroke", m.Stroke)
			w.Str("markText", "markLen", m.Text)
			w.F32("markPx", m.Px)
			w.U8("markBold", weight(m))
			w.U8("markCorner", m.Corner)
			w.F32("markAlpha", m.Alpha)
			w.U8("markDash", flag(m.Dash))
			w.U8("dialAxis", uint8(m.Dial.Axis))
			w.F32("dialStart", m.Dial.Start)
			w.F32("dialTurn", m.Dial.Turn)
			w.F32("dialGhost", m.Dial.Ghost)
			w.F32("dialInc", m.Dial.Inc)
			w.U8("dialLabels", flag(m.Dial.Labels))
		}
		w.U32("caretMark", caret)
		w.U32("caretAt", e.Caret())
		for _, c := range l.Clips {
			w.F32("clipX", c.X)
			w.F32("clipY", c.Y)
			w.F32("clipW", c.W)
			w.F32("clipH", c.H)
		}
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "board_values: %v\n", err)
	}
}

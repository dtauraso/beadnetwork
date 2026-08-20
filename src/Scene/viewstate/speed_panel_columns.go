package viewstate

import (
	"encoding/binary"
	"math"

	"github.com/dtauraso/wirefold/src/Chrome/Panels/SliderPanel"
	B "github.com/dtauraso/wirefold/src/Buffer"
)

func (ui *UIState) writeSpeedPanelColumns(lay SliderPanel.Layout) {
	c := ui.singletonCols
	if c == nil {
		return
	}
	selected := SliderPanel.SelectedIndex(ui.Speed)

	n := len(lay.Ticks)
	xs := make([]byte, 0, n*4)
	ys := make([]byte, 0, n*4)
	ws := make([]byte, 0, n*4)
	hs := make([]byte, 0, n*4)
	sel := make([]byte, 0, n)
	var numText, denText []byte
	numLen := make([]byte, 0, n*4)
	denLen := make([]byte, 0, n*4)

	for i, r := range lay.Ticks {
		xs = binary.LittleEndian.AppendUint32(xs, math.Float32bits(r.X))
		ys = binary.LittleEndian.AppendUint32(ys, math.Float32bits(r.Y))
		ws = binary.LittleEndian.AppendUint32(ws, math.Float32bits(r.W))
		hs = binary.LittleEndian.AppendUint32(hs, math.Float32bits(r.H))

		var on uint8
		if i == selected {
			on = 1
		}
		sel = append(sel, on)

		s := SliderPanel.Settings[i]
		numText = append(numText, s.Num...)
		denText = append(denText, s.Den...)
		numLen = binary.LittleEndian.AppendUint32(numLen, uint32(len(s.Num)))
		denLen = binary.LittleEndian.AppendUint32(denLen, uint32(len(s.Den)))
	}

	c.SetF32(B.ColStreamSpeedPanelBoxX, lay.Box.X)
	c.SetF32(B.ColStreamSpeedPanelBoxY, lay.Box.Y)
	c.SetF32(B.ColStreamSpeedPanelBoxW, lay.Box.W)
	c.SetF32(B.ColStreamSpeedPanelBoxH, lay.Box.H)

	c.SetBytes(B.ColStreamSpeedPanelRectX, xs)
	c.SetBytes(B.ColStreamSpeedPanelRectY, ys)
	c.SetBytes(B.ColStreamSpeedPanelRectW, ws)
	c.SetBytes(B.ColStreamSpeedPanelRectH, hs)
	c.SetBytes(B.ColStreamSpeedPanelSelected, sel)
	c.SetBytes(B.ColStreamSpeedPanelNumText, numText)
	c.SetBytes(B.ColStreamSpeedPanelNumLen, numLen)
	c.SetBytes(B.ColStreamSpeedPanelDenText, denText)
	c.SetBytes(B.ColStreamSpeedPanelDenLen, denLen)

	c.SetF32(B.ColStreamSpeedPanelTrackX, lay.Track.X)
	c.SetF32(B.ColStreamSpeedPanelTrackY, lay.Track.Y)
	c.SetF32(B.ColStreamSpeedPanelTrackW, lay.Track.W)
	c.SetF32(B.ColStreamSpeedPanelTrackH, lay.Track.H)
}

package viewstate

import (
	"encoding/binary"
	"math"

	"github.com/dtauraso/wirefold/nodes/Wiring/tiltpanel"
	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
)

func (ui *UIState) writeTiltPanelColumns(lay tiltpanel.Layout) {
	c := ui.singletonCols
	if c == nil {
		return
	}

	n := len(lay.Columns)
	appendF32 := func(dst []byte, v float32) []byte {
		return binary.LittleEndian.AppendUint32(dst, math.Float32bits(v))
	}

	rows := make([]byte, 0, n*4)
	labelLen := make([]byte, 0, n*4)
	var labelText []byte

	headX := make([]byte, 0, n*4)
	headY := make([]byte, 0, n*4)
	headW := make([]byte, 0, n*4)
	headH := make([]byte, 0, n*4)

	roundsX := make([]byte, 0, n*4)
	roundsY := make([]byte, 0, n*4)
	roundsW := make([]byte, 0, n*4)
	roundsH := make([]byte, 0, n*4)

	msgsX := make([]byte, 0, n*4)
	msgsY := make([]byte, 0, n*4)
	msgsW := make([]byte, 0, n*4)
	msgsH := make([]byte, 0, n*4)

	for _, col := range lay.Columns {
		rows = binary.LittleEndian.AppendUint32(rows, uint32(col.NodeRow))
		labelText = append(labelText, col.Label...)
		labelLen = binary.LittleEndian.AppendUint32(labelLen, uint32(len(col.Label)))

		headX = appendF32(headX, col.Head.X)
		headY = appendF32(headY, col.Head.Y)
		headW = appendF32(headW, col.Head.W)
		headH = appendF32(headH, col.Head.H)

		roundsX = appendF32(roundsX, col.Rounds.X)
		roundsY = appendF32(roundsY, col.Rounds.Y)
		roundsW = appendF32(roundsW, col.Rounds.W)
		roundsH = appendF32(roundsH, col.Rounds.H)

		msgsX = appendF32(msgsX, col.Msgs.X)
		msgsY = appendF32(msgsY, col.Msgs.Y)
		msgsW = appendF32(msgsW, col.Msgs.W)
		msgsH = appendF32(msgsH, col.Msgs.H)
	}

	c.SetF32(B.ColStreamTiltPanelBoxX, lay.Box.X)
	c.SetF32(B.ColStreamTiltPanelBoxY, lay.Box.Y)
	c.SetF32(B.ColStreamTiltPanelBoxW, lay.Box.W)
	c.SetF32(B.ColStreamTiltPanelBoxH, lay.Box.H)

	c.SetF32(B.ColStreamTiltPanelStartX, lay.Start.X)
	c.SetF32(B.ColStreamTiltPanelStartY, lay.Start.Y)
	c.SetF32(B.ColStreamTiltPanelStartW, lay.Start.W)
	c.SetF32(B.ColStreamTiltPanelStartH, lay.Start.H)

	c.SetF32(B.ColStreamTiltPanelResetX, lay.Reset.X)
	c.SetF32(B.ColStreamTiltPanelResetY, lay.Reset.Y)
	c.SetF32(B.ColStreamTiltPanelResetW, lay.Reset.W)
	c.SetF32(B.ColStreamTiltPanelResetH, lay.Reset.H)

	c.SetBytes(B.ColStreamTiltPanelStartText, []byte(tiltpanel.StartLabel))
	c.SetBytes(B.ColStreamTiltPanelResetText, []byte(tiltpanel.ResetLabel))

	c.SetBytes(B.ColStreamTiltPanelColNodeRow, rows)
	c.SetBytes(B.ColStreamTiltPanelColLabelText, labelText)
	c.SetBytes(B.ColStreamTiltPanelColLabelLen, labelLen)

	c.SetBytes(B.ColStreamTiltPanelHeadX, headX)
	c.SetBytes(B.ColStreamTiltPanelHeadY, headY)
	c.SetBytes(B.ColStreamTiltPanelHeadW, headW)
	c.SetBytes(B.ColStreamTiltPanelHeadH, headH)

	c.SetBytes(B.ColStreamTiltPanelRoundsX, roundsX)
	c.SetBytes(B.ColStreamTiltPanelRoundsY, roundsY)
	c.SetBytes(B.ColStreamTiltPanelRoundsW, roundsW)
	c.SetBytes(B.ColStreamTiltPanelRoundsH, roundsH)

	c.SetBytes(B.ColStreamTiltPanelMsgsX, msgsX)
	c.SetBytes(B.ColStreamTiltPanelMsgsY, msgsY)
	c.SetBytes(B.ColStreamTiltPanelMsgsW, msgsW)
	c.SetBytes(B.ColStreamTiltPanelMsgsH, msgsH)
}

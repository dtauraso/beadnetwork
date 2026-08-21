package viewstate

import (
	"encoding/binary"
	"math"

	B "github.com/dtauraso/wirefold/src/Buffer"
	"github.com/dtauraso/wirefold/src/Buffer/colstream"
	"github.com/dtauraso/wirefold/src/Camera"
	"github.com/dtauraso/wirefold/src/Chrome/Panels/PolarRulesPanel"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/AngleDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Pills/NodesDropdown"
	"github.com/dtauraso/wirefold/src/Chrome/Tabs"
)

func (ui *UIState) SetSingletonColumns(set *colstream.ColumnSet) {
	ui.singletonCols = set
}

func (ui *UIState) SetSceneRoot(sceneRoot string) {
	if sceneRoot == "" {
		return
	}
	ui.rulesValues = PolarRulesPanel.NewValueWriter(sceneRoot)
	ui.nodesPillValues = NodesDropdown.NewValueWriter(sceneRoot)
	ui.anglePillValues = AngleDropdown.NewValueWriter(sceneRoot)
	ui.tabStripValues = Tabs.NewValueWriter(sceneRoot)
}

func (ui *UIState) writeSceneColumns() {
	c := ui.singletonCols
	if c == nil {
		return
	}
	c.SetI32(B.ColStreamSceneNodeCount, ui.OwnerCounts.Nodes)
	c.SetI32(B.ColStreamSceneEdgeCount, ui.OwnerCounts.Edges)
}

func writeRingSurfaceColumns(c *colstream.ColumnSet, baseX, baseY, baseZ int, pts []float32) {
	if c == nil || len(pts)%3 != 0 {
		return
	}
	n := len(pts) / 3
	xs := make([]byte, 0, n*4)
	ys := make([]byte, 0, n*4)
	zs := make([]byte, 0, n*4)
	for i := 0; i < n; i++ {
		xs = binary.LittleEndian.AppendUint32(xs, math.Float32bits(pts[i*3]))
		ys = binary.LittleEndian.AppendUint32(ys, math.Float32bits(pts[i*3+1]))
		zs = binary.LittleEndian.AppendUint32(zs, math.Float32bits(pts[i*3+2]))
	}
	c.SetBytes(baseX, xs)
	c.SetBytes(baseY, ys)
	c.SetBytes(baseZ, zs)
}

func (ui *UIState) WriteNodeRingSurfaceColumns(pts []float32) {
	writeRingSurfaceColumns(ui.singletonCols,
		B.ColStreamNodeRingPointX, B.ColStreamNodeRingPointY, B.ColStreamNodeRingPointZ, pts)
}

func (ui *UIState) WriteBeadRingSurfaceColumns(pts []float32) {
	writeRingSurfaceColumns(ui.singletonCols,
		B.ColStreamBeadRingPointX, B.ColStreamBeadRingPointY, B.ColStreamBeadRingPointZ, pts)
}

func (ui *UIState) FovDeg() float64 {
	return Camera.FovDegForHeight(ui.ViewH)
}

func (ui *UIState) writePointerTargetColumns() {
	c := ui.singletonCols
	if c == nil {
		return
	}
	t := ui.Pointer
	c.SetF32(B.ColStreamPointerTargetX, t.Rect.X)
	c.SetF32(B.ColStreamPointerTargetY, t.Rect.Y)
	c.SetF32(B.ColStreamPointerTargetW, t.Rect.W)
	c.SetF32(B.ColStreamPointerTargetH, t.Rect.H)
	c.SetU8(B.ColStreamPointerTargetKind, uint8(t.Kind))

	tx, ty, _, _ := t.TipRect(float32(ui.ViewW))
	c.SetF32(B.ColStreamPointerTargetTipX, tx)
	c.SetF32(B.ColStreamPointerTargetTipY, ty)
	c.SetBytes(B.ColStreamPointerTargetTipText, []byte(t.Tip))
}


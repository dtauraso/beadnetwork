package viewstate

import (
	"github.com/dtauraso/wirefold/nodes/Wiring/geom/polar"
	B "github.com/dtauraso/wirefold/tools/topology-vscode/Buffer"
	"github.com/dtauraso/wirefold/tools/topology-vscode/Buffer/colstream"
)

func (ui *UIState) SetSingletonColumns(set *colstream.ColumnSet) {
	ui.singletonCols = set
}

func (ui *UIState) writeSceneColumns(sc polar.SceneSphere) {
	c := ui.singletonCols
	if c == nil {
		return
	}
	c.SetF32(B.ColStreamSceneCX, float32(sc.Center.X))
	c.SetF32(B.ColStreamSceneCY, float32(sc.Center.Y))
	c.SetF32(B.ColStreamSceneCZ, float32(sc.Center.Z))
	c.SetF32(B.ColStreamSceneRadius, float32(sc.Radius))
	c.SetF32(B.ColStreamSceneConstantR, float32(ui.Constants.ConstantR))
	c.SetI32(B.ColStreamSceneMaxIndexPhi, int32(ui.Constants.MaxIndexPhi))
	c.SetI32(B.ColStreamSceneMaxIndexTheta, int32(ui.Constants.MaxIndexTheta))
}

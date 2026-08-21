package PolarRulesPanel

import "github.com/dtauraso/wirefold/src/valuefile"

const ValueRelDir = "view/chrome/rules-panel"

var PanelValueNames = []string{
	"clipY", "clipH", "scrollY",
	"boxX", "boxY", "boxW", "boxH",
	"open",
	"toggleX", "toggleY", "toggleH", "toggleText",
	"rowKind",
	"rowX", "rowY", "rowW", "rowH",
	"rowTextData", "rowTextLen",
	"rowGlyphData", "rowGlyphLen",
	"rowFree",
	"rowNodeRow", "rowEdgeRow",
	"rowCheck", "rowCheckX", "rowCheckY", "rowCheckW", "rowCheckH",
	"rowValue", "rowValueX", "rowValueY",
	"rowSharedX", "rowSharedY", "rowSharedW", "rowSharedH",
	"rowEditing",
	"draftText", "draftX", "draftY", "draftW", "draftH",
	"menuOpen", "menuAnchorRow",
	"menuX", "menuY", "menuW", "menuH",
	"menuRowX", "menuRowY", "menuRowW", "menuRowH",
	"menuCheckX", "menuCheckY",
	"menuLabelData", "menuLabelLen",
	"menuNodeRow",
}

func ValueRelPath(name string) string {
	return ValueRelDir + "/" + valuefile.BlobFileName(name)
}

func ValuePath(sceneRoot, name string) string {
	for _, n := range PanelValueNames {
		if n == name {
			return sceneRoot + "/" + ValueRelPath(name)
		}
	}
	panic("PolarRulesPanel.ValuePath: " + name + " is not a panel value, so Go and the renderer disagree about where it lives; add it to PanelValueNames in src/Chrome/Panels/PolarRulesPanel/panel_values.go")
}

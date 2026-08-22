package PolarRulesPanel

const ValueRelFile = "view/chrome/rules-panel.bin"

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

func ValueRelPath() string { return ValueRelFile }

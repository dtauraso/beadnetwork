package Pills

import (
	"slices"

	"github.com/dtauraso/beadnetwork/Categories/Chrome/Panels/Sheet"
	Flags "github.com/dtauraso/beadnetwork/Categories/Scene/View/Flags"
)

const OpOverlay = "overlay"

var dockRows = []struct {
	Name  string
	Flags []string
}{
	{"node", []string{"nodeBody", "nodeRing", "ringPick"}},
	{"state", []string{"selectionRing", "hoverRing"}},
	{"node poles", []string{"nodePoles", "nodePoleSphere"}},
	{"scene poles", []string{"scenePoles", "allPoleSpheres"}},
	{"guides", []string{"tori", "handholds"}},
	{"vectors", []string{"sceneVectors", "edgeVectors", "ruleChannels"}},
	{"labels", []string{"labelsGlobal"}},
}

func itemOf(flag string, groups []Group) (Item, bool) {
	for _, g := range groups {
		for _, it := range g.Items {
			if it.Flag == flag {
				return it, true
			}
		}
		if it, ok := itemOf(flag, g.Groups); ok {
			return it, true
		}
	}
	return Item{}, false
}

func flagOn(ov *Flags.OverlayState, flag string) bool {
	read, ok := Flags.OverlayFlagRead[flag]
	return ok && read(ov)
}

func DockPill(ov *Flags.OverlayState, hidden []string) Sheet.Pill {
	active := flagOn(ov, GuidelinesFlag)
	all := "overlays off"
	if active {
		all = "overlays on"
	}
	rows := [][]Sheet.Cell{{Sheet.Button(all, Sheet.Act{Op: OpOverlay, K: GuidelinesFlag}, active)}}
	for _, r := range dockRows {
		row := []Sheet.Cell{Sheet.Text(r.Name)}
		for _, f := range r.Flags {
			it, ok := itemOf(f, Tree)
			if !ok || slices.Contains(hidden, f) {
				continue
			}
			on := flagOn(ov, f)
			icon := it.Icon
			if f == "labelsGlobal" {
				icon = LabelsIcon(on)
			}
			row = append(row, Sheet.Button(icon+" "+it.Label, Sheet.Act{Op: OpOverlay, K: f}, on).Dimmed(!active))
		}
		if len(row) > 1 {
			rows = append(rows, row)
		}
	}
	return Sheet.Pill{Label: Label, Bottom: true, Groups: []Sheet.Group{
		{Title: "overlays — what's drawn over the diagram", Rows: rows},
	}}
}

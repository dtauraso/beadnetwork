package Pills

type Item struct {
	Flag  string
	Icon  string
	Label string
}

type Group struct {
	Heading string

	Panel string

	Items  []Item
	Groups []Group
}

const GuidelinesFlag = "overlays"

const Label = "Overlays"

var Tree = []Group{
	{
		Heading: "NODE",
		Panel:   "node",
		Groups: []Group{
			{
				Heading: "SHAPE",
				Panel:   "nodeShape",
				Items: []Item{
					{Flag: "nodeBody", Icon: "●", Label: "body"},
					{Flag: "nodeRing", Icon: "○", Label: "ring"},
					{Flag: "ringPick", Icon: "◌", Label: "ring band"},
				},
			},
			{
				Heading: "STATE",
				Panel:   "nodeState",
				Items: []Item{
					{Flag: "selectionRing", Icon: "◉", Label: "selection"},
					{Flag: "hoverRing", Icon: "◍", Label: "hover"},
				},
			},
			{
				Heading: "POLES",
				Panel:   "nodePoles",
				Items: []Item{
					{Flag: "nodePoles", Icon: "⊹", Label: "node poles"},
					{Flag: "nodePoleSphere", Icon: "◍", Label: "pole sphere"},
				},
			},
		},
	},
	{
		Heading: "SCENE",
		Panel:   "scene",
		Groups: []Group{
			{
				Heading: "GUIDES",
				Panel:   "sceneGuides",
				Items: []Item{
					{Flag: "tori", Icon: "◎", Label: "rings"},
					{Flag: "handholds", Icon: "⊙", Label: "grips"},
				},
			},
			{
				Heading: "POLES",
				Panel:   "scenePoles",
				Items: []Item{
					{Flag: "scenePoles", Icon: "⊹", Label: "scene poles"},
					{Flag: "allPoleSpheres", Icon: "◍", Label: "all pole spheres"},
				},
			},
			{
				Heading: "VECTORS",
				Panel:   "sceneVectors",
				Items: []Item{
					{Flag: "sceneVectors", Icon: "↗", Label: "scene vectors"},
					{Flag: "ruleChannels", Icon: "⇄", Label: "rule channels"},
				},
			},
			{
				Heading: "LABELS",
				Panel:   "sceneLabels",
				Items: []Item{
					{Flag: "labelsGlobal", Icon: "▴", Label: "labels"},
				},
			},
		},
	},
}

func LabelsIcon(on bool) string {
	if on {
		return "▴"
	}
	return "▾"
}

func allFlags(g Group) []string {
	out := make([]string, 0, len(g.Items))
	for _, it := range g.Items {
		out = append(out, it.Flag)
	}
	for _, sub := range g.Groups {
		out = append(out, allFlags(sub)...)
	}
	return out
}

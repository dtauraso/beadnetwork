package Wiring

import (
	"strings"
	"testing"

	T "github.com/dtauraso/wirefold/Trace"
)

// overlay_gen_test.go — independent behavior check for the generated overlay
// Toggle wiring in overlay_gen.go. This is the trust frontier of the generator:
// it exercises the mechanical name-derivation (flag → field / method) that
// replaced the hand-written methods. For EACH overlay flag it asserts that
// Toggle flips the owned overlayState bool. The RowEvent this toggle implies is
// written by the caller (stdin_reader.go's applyUpdate), not by overlayState
// itself, so it is out of scope here. The exception flags (scene/node poles
// breadcrumb) are covered explicitly.

// overlayCase describes one generated overlay flag by its behavior, using closures
// so the test reads/writes the concrete overlayState field without reflection.
type overlayCase struct {
	name   string
	get    func(*overlayState) bool
	toggle func(*overlayState, *T.Trace)
	crumb  string // non-empty => breadcrumb node arg expected on Toggle (poles)
}

var overlayCases = []overlayCase{
	{
		name:   "tori",
		get:    func(o *overlayState) bool { return o.sceneToriVisible },
		toggle: (*overlayState).ToggleSceneTori,
	},
	{
		name:   "scenePoles",
		get:    func(o *overlayState) bool { return o.scenePolesVisible },
		toggle: (*overlayState).ToggleScenePoles,
		crumb:  "scene",
	},
	{
		name:   "nodePoles",
		get:    func(o *overlayState) bool { return o.nodePolesVisible },
		toggle: (*overlayState).ToggleNodePoles,
		crumb:  "nodes",
	},
	{
		name:   "selSpherePoles",
		get:    func(o *overlayState) bool { return o.selSpherePolesVisible },
		toggle: (*overlayState).ToggleSelSpherePoles,
	},
	{
		name:   "handholds",
		get:    func(o *overlayState) bool { return o.handholdsVisible },
		toggle: (*overlayState).ToggleHandholds,
	},
	{
		name:   "labelsGlobal",
		get:    func(o *overlayState) bool { return o.labelsGlobalVisible },
		toggle: (*overlayState).ToggleLabelsGlobal,
	},
	{
		name:   "overlays",
		get:    func(o *overlayState) bool { return o.overlaysVisible },
		toggle: (*overlayState).ToggleOverlaysVis,
	},
}

func TestOverlayToggleFlips(t *testing.T) {
	for _, c := range overlayCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			// Start from a known false, so a flip lands on true.
			var o overlayState
			var dbg strings.Builder
			tr := T.New(0)
			tr.SetSink(&dbg)
			c.toggle(&o, tr)

			if !c.get(&o) {
				t.Fatalf("%s: Toggle did not flip field false->true", c.name)
			}
			if c.crumb != "" {
				s := dbg.String()
				if !strings.Contains(s, "pole-toggle-go") || !strings.Contains(s, `"node":"`+c.crumb+`"`) {
					t.Fatalf("%s: Toggle did not emit breadcrumb for scope %q; sink=%q", c.name, c.crumb, s)
				}
			}
		})
	}
}

// TestDefaultOverlayState pins the startup snapshot: 7 flags default ON, and
// doubleLinks is the single exception that defaults OFF (overlay_gen.go's
// doubleLinks:{defaultOff:true}). Without the OFF assertion below, a generator bug
// flipping that default would produce self-consistent output that passes both
// check-generated and this test.
func TestDefaultOverlayState(t *testing.T) {
	d := defaultOverlayState()
	on := []bool{
		d.sceneToriVisible, d.scenePolesVisible, d.nodePolesVisible,
		d.selSpherePolesVisible, d.handholdsVisible,
		d.labelsGlobalVisible, d.overlaysVisible,
	}
	for i, v := range on {
		if !v {
			t.Fatalf("defaultOverlayState field #%d should default ON", i)
		}
	}
	if d.doubleLinksVisible {
		t.Fatal("defaultOverlayState: doubleLinksVisible should default OFF (defaultOff:true in overlay_gen.go)")
	}
}
